package handler

import (
	"context"
	"encoding/xml"
	"log"
	"tally-connector/api/middlewares"
	"tally-connector/config"
	"tally-connector/internal/db"
	"tally-connector/internal/loader"
	"tally-connector/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

type TotalBalance struct {
	LedgerName     string  `db:"ledger"`
	TotalNetDebit  float64 `db:"total_net_debit" json:"total_net_debit"`
	TotalNetCredit float64 `db:"total_net_credit" json:"total_net_credit"`
	OpeningBalance float64 `db:"opening_balance"`
	ClosingBalance float64 `db:"closing_balance"`
}

func getOpeningBalance(ctx context.Context, ledger_ids []string) (map[string]float64, error) {
	openingBalances := make(map[string]float64)

	for _, ledgerID := range ledger_ids {
		var openingBalance float64
		err := db.GetDB().QueryRow(ctx, "SELECT opening_balance FROM mst_ledger WHERE name = $1", ledgerID).Scan(&openingBalance)
		if err != nil {
			return nil, err
		}
		openingBalances[ledgerID] = openingBalance
	}

	return openingBalances, nil
}
func GetTotal(id string) {
	result, err := getTotalBalance(context.Background(), TotalBalanceParams{
		StartDate: "2025-01-01",
		EndDate:   "2025-12-31",
		LedgerIDs: []string{id},
	})
	if err != nil {
		log.Println("Error fetching total balance:", err)
		return
	}
	log.Println("Total balance for", id, ":", result)
}

type TotalBalanceParams struct {
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	LedgerIDs    []string `json:"ledger_ids"`
	SkipVouchers []string `json:"skip_vouchers"`
}

func getTotalBalance(ctx context.Context, params TotalBalanceParams) (map[string]TotalBalance, error) {

	var start_date = params.StartDate
	var end_date = params.EndDate
	var ledger_ids = params.LedgerIDs

	var totalBalances = make(map[string]TotalBalance)

	if start_date == "" || end_date == "" {
		return totalBalances, nil
	}

	query := `
WITH date_range AS (
    SELECT ?::date AS start_date, ?::date AS end_date
)
SELECT
    a.ledger,
    COALESCE(SUM(CASE WHEN t.date < dr.start_date THEN a.amount END), 0) AS opening_balance,
    COALESCE(SUM(CASE WHEN t.date <= dr.end_date THEN a.amount END), 0) AS closing_balance,
    COALESCE(SUM(CASE WHEN t.date BETWEEN dr.start_date AND dr.end_date AND a.amount > 0 THEN a.amount END), 0) AS total_net_debit,
    COALESCE(SUM(CASE WHEN t.date BETWEEN dr.start_date AND dr.end_date AND a.amount < 0 THEN a.amount END), 0) AS total_net_credit
FROM
    trn_accounting a
    JOIN trn_voucher t ON a.guid = t.guid
    CROSS JOIN date_range dr
WHERE
    a.ledger = ANY(?)
`

	args := []interface{}{start_date, end_date, pq.Array(ledger_ids)}

	if params.SkipVouchers != nil {
		query += ` AND t.guid  <> ALL(?)`
		args = append(args, pq.Array(params.SkipVouchers))
	}

	query += `
		GROUP BY
    a.ledger
		ORDER BY
    opening_balance DESC
	`

	stmt := psql.Raw(query, args...)

	var totalBalance = []TotalBalance{}

	err := pgxscan.Select(ctx, db.GetDB(), &totalBalance, stmt.String(), args...)

	if err != nil {
		return totalBalances, err
	}

	// fetch initial opening balance
	openingBalances, err := getOpeningBalance(ctx, ledger_ids)
	if err != nil {
		return totalBalances, err
	}

	for idx, balance := range totalBalance {
		if initial, ok := openingBalances[balance.LedgerName]; ok {
			totalBalance[idx].OpeningBalance = initial + balance.OpeningBalance
			totalBalance[idx].ClosingBalance = initial + balance.ClosingBalance
		}

		totalBalances[balance.LedgerName] = totalBalance[idx]
	}

	// loop on ledger ids.. if it's not in totalBalances, initialize it
	for _, ledgerID := range ledger_ids {
		// if it's not in totalBalances
		if _, ok := totalBalances[ledgerID]; ok {
			continue
		}

		// assign opening and closing balance to initial opening balance
		if _, ok := openingBalances[ledgerID]; ok {
			totalBalances[ledgerID] = TotalBalance{
				LedgerName:     ledgerID,
				OpeningBalance: openingBalances[ledgerID],
				ClosingBalance: openingBalances[ledgerID],
			}
		}

	}

	return totalBalances, nil
}

// func getLedgerDetails(ctx context.Context, ledgerID string) (models.Ledger, error) {
// 	var ledger = models.Ledger{}
// 	err := db.GetDB().QueryRow(ctx, "SELECT * FROM mst_ledger WHERE name = $1", ledgerID).Scan(&ledger)
// 	if err != nil {
// 		return models.Ledger{}, err
// 	}

// 	if ledger.Name != "" {
// 		totalBalance, _ := getTotalBalance(ctx, []string{ledger.Name})

// 		ledger.TotalNetCredit = totalBalance[ledger.Name].TotalNetCredit
// 		ledger.TotalNetDebit = totalBalance[ledger.Name].TotalNetDebit

// 		ledger.OpeningBalance = totalBalance[ledger.Name].OpeningBalance
// 		ledger.ClosingBalance = totalBalance[ledger.Name].ClosingBalance

// 	}

// 	return ledger, nil
// }

func FetchLedgerHandler(c *gin.Context) {
	var queryParams middlewares.PaginationParams

	err := c.BindQuery(&queryParams)
	if err != nil {
		log.Println("Error in binding query ", err)
		c.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid query parameters",
		})
		return
	}

	// if queryParams.StartDate == "" || queryParams.EndDate == "" {
	// 	log.Println("Start date or end date is missing")
	// 	c.JSON(400, gin.H{
	// 		"error":   "Start date or end date is missing",
	// 		"message": "Please provide start_date and end_date in query parameters",
	// 	})
	// 	return
	// }

	var ledgers = []models.Ledger{}

	var ctx = context.Background()

	q := psql.Select(
		sm.Columns("name", "parent", "alias", "opening_balance", "closing_balance"),
		sm.From("mst_ledger"),
		sm.OrderBy("name ASC"),
		sm.Offset(queryParams.Page*queryParams.Limit),
		sm.Limit(queryParams.Limit),
	)

	var searchFilter = sm.Where(psql.Or(
		psql.Quote("name").ILike(psql.Arg("%"+queryParams.Search+"%")),
		psql.Quote("alias").ILike(psql.Arg("%"+queryParams.Search+"%")),
	))

	if queryParams.Search != "" {
		q.Apply(searchFilter)
	}

	query, args, err := q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Select(ctx, db.GetDB(), &ledgers, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve ledgers",
		})
		return
	}

	// count
	var count int64
	countQuery := psql.Select(
		sm.Columns("COUNT(*)"),
		sm.From("mst_ledger"),
	)

	if queryParams.Search != "" {
		countQuery.Apply(searchFilter)
	}

	query, args, err = countQuery.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Get(ctx, db.GetDB(), &count, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve ledgers",
		})
		return
	}

	// total balances
	ledgerIds := make([]string, 0, len(ledgers))
	for _, ledger := range ledgers {
		ledgerIds = append(ledgerIds, ledger.Name)
	}

	totalBalances, err := getTotalBalance(ctx, TotalBalanceParams{
		StartDate: queryParams.StartDate,
		EndDate:   queryParams.EndDate,
		LedgerIDs: ledgerIds,
	})
	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve total balances",
		})
		return
	}

	for i, ledger := range ledgers {
		if balance, ok := totalBalances[ledger.Name]; ok {
			ledger.OpeningBalance = balance.OpeningBalance
			ledger.ClosingBalance = balance.ClosingBalance
			ledger.TotalNetCredit = balance.TotalNetCredit
			ledger.TotalNetDebit = balance.TotalNetDebit
		}

		ledgers[i] = ledger
	}

	c.JSON(200, gin.H{"ledgers": ledgers, "count": count})

}

func FetchLedgerAutoComplete(c *gin.Context) {
	ctx := c.Request.Context()

	var ledgers []models.Ledger

	q := psql.Select(
		sm.Columns("name", "alias"),
		sm.From("mst_ledger"),
		sm.OrderBy("name ASC"),
	)

	query, args, err := q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Select(ctx, db.GetDB(), &ledgers, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve ledgers",
		})
		return
	}

	c.JSON(200, gin.H{"ledgers": ledgers})

}

func AlterLedger(c *gin.Context) {
	type RequestBody struct {
		Guid        string `json:"guid" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Alias       string `json:"alias"`
		Description string `json:"description"`
	}

	var reqBody RequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid request body",
		})
		return
	}

	c.JSON(200, gin.H{"message": "Ledger updated successfully"})
}

func CreateLedgerHandler(c *gin.Context) {
	type RequestBody struct {
		Name       string `json:"name" binding:"required"`
		Alias      string `json:"alias" binding:"required"`
		Parent     string `json:"parent" binding:"required"`
		IsPositive string `json:"is_positive" binding:"required"`
	}

	var reqBody RequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid request body",
		})
		return
	}
	ctx := c.Request.Context()

	type Envelope struct {
		HEADER struct {
			VERSION      string `json:"VERSION" xml:"VERSION"`
			TALLYREQUEST string `json:"TALLYREQUEST" xml:"TALLYREQUEST"`
			TYPE         string `json:"TYPE" xml:"TYPE"`
			ID           string `json:"ID" xml:"ID"`
		} `json:"HEADER" xml:"HEADER"`
		BODY struct {
			DESC struct {
				STATICVARIABLES struct {
					SVCURRENTCOMPANY string `json:"SVCURRENTCOMPANY" xml:"SVCURRENTCOMPANY"`
				} `json:"STATICVARIABLES" xml:"STATICVARIABLES"`
			} `json:"DESC" xml:"DESC"`
			DATA struct {
				TALLYMESSAGE struct {
					LEDGER struct {
						PARENT           string `json:"PARENT" xml:"PARENT"`
						ISDEEMEDPOSITIVE string `json:"ISDEEMEDPOSITIVE" xml:"ISDEEMEDPOSITIVE"`
						NAMELIST         struct {
							NAME []string `json:"NAME" xml:"NAME"`
							TYPE string   `json:"-" xml:"TYPE,attr"`
						} `json:"NAME.LIST" xml:"NAME.LIST"`
						ACTION string `json:"-" xml:"ACTION,attr"`
					} `json:"LEDGER" xml:"LEDGER"`
				} `json:"TALLYMESSAGE" xml:"TALLYMESSAGE"`
			} `json:"DATA" xml:"DATA"`
		} `json:"BODY" xml:"BODY"`
	}

	config, _ := config.LoadConfig()

	var env Envelope
	env.HEADER.VERSION = "1"
	env.HEADER.TALLYREQUEST = "Import"
	env.HEADER.TYPE = "Data"
	env.HEADER.ID = "All Masters"
	env.BODY.DESC.STATICVARIABLES.SVCURRENTCOMPANY = config.Tally.Company
	env.BODY.DATA.TALLYMESSAGE.LEDGER.PARENT = reqBody.Parent
	env.BODY.DATA.TALLYMESSAGE.LEDGER.NAMELIST.NAME = []string{reqBody.Name, reqBody.Alias}
	env.BODY.DATA.TALLYMESSAGE.LEDGER.ISDEEMEDPOSITIVE = reqBody.IsPositive
	env.BODY.DATA.TALLYMESSAGE.LEDGER.NAMELIST.TYPE = "String"
	env.BODY.DATA.TALLYMESSAGE.LEDGER.ACTION = "Create"

	marshal, err := xml.Marshal(env)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to marshal XML",
		})
		return
	}

	result, err := loader.CallTallyApi(ctx, marshal)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to call Tally API",
		})
		return
	}

	log.Println("Tally API Response:", string(result))
	loader.ImportAll("mst_ledger")

	c.JSON(200, gin.H{"message": "Ledger created successfully"})

}
/*
	1.


	2.
*/
