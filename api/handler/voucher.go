package handler

import (
	"context"
	"log"
	"tally-connector/api/middlewares"
	"tally-connector/internal/db"
	"tally-connector/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func getNameBySearch(ctx context.Context, ledgerName string) (string, error) {
	if ledgerName == "" {
		return "", nil
	}
	var name string

	err := db.GetDB().QueryRow(ctx, "SELECT name FROM mst_ledger WHERE name = $1 OR alias = $1", ledgerName).Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil
}

type VoucherQueryParams struct {
	middlewares.PaginationParams
	LedgerName  string `form:"ledger_name"`
	VoucherType string `form:"voucher_type"`
}

func FetchVoucherHandler(c *gin.Context) {
	var queryParams VoucherQueryParams
	err := c.ShouldBindQuery(&queryParams)
	if err != nil {
		log.Println("Error in binding query ", err)
		c.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid query parameters",
		})
		return
	}

	var ctx = c.Request.Context()

	name, err := getNameBySearch(ctx, queryParams.LedgerName)
	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve ledger name",
		})
		return
	}

	var vouchers = []models.TrnVoucher{}
	var count int64

	q := psql.Select(
		sm.Columns("guid", "date", "party_name", "voucher_type", "voucher_number", "narration"),
		sm.From("trn_voucher"),
		sm.Limit(queryParams.Limit),
		sm.Offset(queryParams.Page*queryParams.Limit),
		sm.OrderBy("date ASC"),
	)

	countQuery := psql.Select(
		sm.Columns("COUNT(*)"),
		sm.From("trn_voucher"),
	)

	if name != "" {
		var nameFilter = sm.Where(psql.Quote("party_name").EQ(psql.Arg(name)))
		q.Apply(nameFilter)
		countQuery.Apply(nameFilter)
	}

	if queryParams.Search != "" {
		searchFilter := sm.Where(
			psql.Or(
				psql.Quote("narration").ILike(psql.Arg("%"+queryParams.Search+"%")),
				psql.Quote("party_name").ILike(psql.Arg("%"+queryParams.Search+"%")),
			),
		)
		q.Apply(searchFilter)
		countQuery.Apply(searchFilter)
	}

	if queryParams.StartDate != "" && queryParams.EndDate != "" {
		dateFilter := sm.Where(
			psql.And(
				psql.Quote("date").GTE(psql.Arg(queryParams.StartDate)),
				psql.Quote("date").LTE(psql.Arg(queryParams.EndDate)),
			),
		)
		q.Apply(dateFilter)
		countQuery.Apply(dateFilter)
	}

	if queryParams.VoucherType != "" {
		var voucherTypeFilter = sm.Where(psql.Quote("voucher_type").EQ(psql.Arg(queryParams.VoucherType)))
		q.Apply(voucherTypeFilter)
		countQuery.Apply(voucherTypeFilter)
	}

	query, args, err := q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Select(ctx, db.GetDB(), &vouchers, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve vouchers",
		})
		return
	}

	val, args, err := countQuery.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build count query",
		})
		return
	}

	err = db.GetDB().QueryRow(ctx, val, args...).Scan(&count)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve voucher count",
		})
		return
	}

	c.JSON(200, gin.H{
		"count":    count,
		"vouchers": vouchers,
	})

}

func FetchVoucherDetailsHandler(c *gin.Context) {

	var ctx = c.Request.Context()
	var ledgerID = c.Param("id")

	var voucher = models.TrnVoucher{}
	var inventory = []models.TrnInventory{}
	var inventory_accounting = []models.TrnInventoryAccounting{}
	var accounting = []models.TrnAccounting{}
	// Transaction Inventory
	q := psql.Select(
		sm.Columns("item", "quantity", "rate", "amount", "additional_amount", "discount_amount"),
		sm.From("trn_inventory"),
		sm.Where(psql.Quote("guid").EQ(psql.Arg(ledgerID))),
	)

	query, args, err := q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Select(ctx, db.GetDB(), &inventory, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve inventory",
		})
		return
	}

	// trn_inventory_accounting
	q = psql.Select(
		sm.Columns("guid", "ledger", "amount", "additional_allocation_type"),
		sm.From("trn_inventory_accounting"),
		sm.Where(psql.Quote("guid").EQ(psql.Arg(ledgerID))),
	)

	query, args, err = q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Select(ctx, db.GetDB(), &inventory_accounting, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve inventory accounting",
		})
		return
	}

	// voucher itself
	q = psql.Select(
		sm.Columns("guid", "date", "party_name", "voucher_type", "voucher_number", "narration"),
		sm.From("trn_voucher"),
		sm.Where(psql.Quote("guid").EQ(psql.Arg(ledgerID))),
	)

	query, args, err = q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Get(
		ctx,
		db.GetDB(),
		&voucher,
		query,
		args...,
	)

	if err != nil {

		if pgxscan.NotFound(err) {
			c.JSON(404, gin.H{
				"error":   "Voucher Not Found",
				"message": "Voucher not found",
			})
			return
		}

		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve voucher",
		})
		return
	}

	// accounting
	q = psql.Select(
		sm.Columns("guid", "ledger", "amount"),
		sm.From("trn_accounting"),
		sm.Where(psql.Quote("guid").EQ(psql.Arg(ledgerID))),
	)

	query, args, err = q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Select(ctx, db.GetDB(), &accounting, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve accounting",
		})
		return
	}

	c.JSON(200, gin.H{
		"voucher":              voucher,
		"inventory":            inventory,
		"inventory_accounting": inventory_accounting,
		"accounting":           accounting,
		// "bills":                bills,
		// "batch":                batch,
		// "bank":                 bank,
	})
}
