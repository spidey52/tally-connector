package handler

import (
	"math"
	"tally-connector/api/middlewares"
	"tally-connector/internal/db"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

type Daybook struct {
	Guid             string    `json:"voucher_guid" db:"voucher_guid"`
	Date             time.Time `json:"voucher_date" db:"voucher_date"`
	VoucherNumber    string    `json:"voucher_number" db:"voucher_number"`
	VoucherType      string    `json:"voucher_type" db:"voucher_type"`
	PartyName        string    `json:"party_name" db:"party_name"`
	AccountingAmount float64   `json:"accounting_amount" db:"accounting_amount"`

	DebitAmount  float64 `json:"debit_amount" db:"-"`
	CreditAmount float64 `json:"credit_amount" db:"-"`
}

func FetchDaybook(c *gin.Context) {
	var ctx = c.Request.Context()
	var paginationParams = c.MustGet("pagination_params").(middlewares.PaginationParams)

	if paginationParams.StartDate == "" || paginationParams.EndDate == "" {

		// fetch last date from voucher

		var lastDate time.Time
		err := db.GetDB().QueryRow(ctx, "SELECT MAX(date) FROM trn_voucher").Scan(&lastDate)

		if err != nil {
			c.JSON(500, gin.H{
				"error":   err.Error(),
				"message": "Failed to retrieve last voucher date",
			})
			return
		}

		paginationParams.StartDate = lastDate.Format("2006-01-02")
		paginationParams.EndDate = lastDate.Format("2006-01-02")

	}

	q := psql.Select(
		sm.Columns("v.guid as voucher_guid", "v.date as voucher_date", "v.voucher_number", "v.voucher_type", "v.party_name", "a.amount as accounting_amount"),
		sm.From("trn_voucher v"),
		sm.LeftJoin("trn_accounting a ON v.guid = a.guid"),
		sm.Where(psql.Quote("v", "party_name").EQ(psql.Quote("a", "ledger"))),
		sm.Where(psql.Quote("v", "date").GTE(psql.Arg(paginationParams.StartDate)).And(psql.Quote("v", "date").LTE(psql.Arg(paginationParams.EndDate)))),
		sm.OrderBy("v.guid ASC"),
		sm.Limit(paginationParams.Limit),
		sm.Offset(paginationParams.Page*paginationParams.Limit),
	)

	var daybook = []Daybook{}

	query, args, err := q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Select(ctx, db.GetDB(), &daybook, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve daybook",
		})
		return
	}

	for i := range daybook {
		isPositive := daybook[i].AccountingAmount >= 0

		absAmount := math.Abs(daybook[i].AccountingAmount)

		if isPositive {
			daybook[i].CreditAmount = absAmount
		} else {
			daybook[i].DebitAmount = absAmount
		}

	}

	countQuery := psql.Select(
		sm.Columns("COUNT(*)"),
		sm.From("trn_voucher v"),
		sm.LeftJoin("trn_accounting a ON v.guid = a.guid"),
		sm.Where(psql.Quote("v", "party_name").EQ(psql.Quote("a", "ledger"))),
		sm.Where(psql.Quote("v", "date").GTE(psql.Arg(paginationParams.StartDate)).And(psql.Quote("v", "date").LTE(psql.Arg(paginationParams.EndDate)))),
	)

	query, args, err = countQuery.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build count query",
		})
		return
	}

	var count int64
	err = db.GetDB().QueryRow(ctx, query, args...).Scan(&count)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve daybook count",
		})
		return
	}

	c.JSON(200, gin.H{
		"count":   count,
		"daybook": daybook,
	})

}
