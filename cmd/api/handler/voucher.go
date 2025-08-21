package handler

import (
	"context"
	"tally-connector/cmd/db"
	"tally-connector/cmd/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func getNameBySearch(ledgerName string) (string, error) {
	var name string
	var ctx = context.Background()

	err := db.GetDB().QueryRow(ctx, "SELECT name FROM mst_ledger WHERE name = $1 OR alias = $1", ledgerName).Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil
}

func FetchVoucherHandler(c *gin.Context) {
	var ledger_name = c.DefaultQuery("ledger_name", "")

	name, err := getNameBySearch(ledger_name)
	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve ledger name",
		})
		return
	}

	var vouchers []models.TrnVoucher
	var count int64

	q := psql.Select(
		sm.Columns("guid", "date", "party_name", "voucher_type", "voucher_number", "narration"),
		sm.From("trn_voucher"),
		sm.Limit(10),
	)

	countQuery := psql.Select(
		sm.Columns("COUNT(*)"),
		sm.From("trn_voucher"),
	)

	var nameFilter = sm.Where(psql.Quote("party_name").EQ(psql.Arg(name)))

	if name != "" {
		q.Apply(nameFilter)
		countQuery.Apply(nameFilter)
	}

	query, args, err := q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Select(context.Background(), db.GetDB(), &vouchers, query, args...)

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

	err = db.GetDB().QueryRow(context.Background(), val, args...).Scan(&count)

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

func FetchLedgerHandler(c *gin.Context) {
	type Ledger struct {
		Name   string `db:"name" json:"name"`
		Parent string `db:"parent" json:"parent"`
		Alias  string `db:"alias" json:"alias"`
	}

	var ledgers []Ledger

	var ctx = context.Background()

	q := psql.Select(
		sm.Columns("name", "parent", "alias"),
		sm.From("mst_ledger"),
		sm.OrderBy("name ASC"),
		sm.Offset(1000),
		sm.Limit(10),
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

func FetchVoucherDetailsHandler(c *gin.Context) {

	var ctx = context.Background()
	var ledgerID = c.Param("id")

	var voucher = models.TrnVoucher{}
	var inventory = []models.TrnInventory{}
	var inventory_accounting = []models.TrnInventoryAccounting{}
	var accounting = []models.TrnAccounting{}
	var bills = []models.TrnBill{}
	var batch = []models.TrnBatch{}
	var bank = []models.TrnBank{}

	// Inventory Accounting
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

	// again overwrite the query
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

	// tans bills

	q = psql.Select(
		sm.Columns("guid", "ledger", "name", "amount", "billtype", "bill_credit_period"),
		sm.From("trn_bill"),
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

	err = pgxscan.Select(ctx, db.GetDB(), &bills, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve bills",
		})
		return
	}

	// tran batch
	q = psql.Select(
		sm.Columns("guid", "item", "name", "quantity", "amount", "godown"),
		sm.From("trn_batch"),
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

	err = pgxscan.Select(ctx, db.GetDB(), &batch, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve batch",
		})
		return
	}

	// tran bank
	q = psql.Select(
		sm.Columns("guid", "ledger", "transaction_type", "instrument_date", "instrument_number", "bank_name", "amount", "bankers_date"),
		sm.From("trn_bank"),
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

	err = pgxscan.Select(ctx, db.GetDB(), &bank, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve bank",
		})
		return
	}

	c.JSON(200, gin.H{
		"voucher":              voucher,
		"inventory":            inventory,
		"inventory_accounting": inventory_accounting,
		"accounting":           accounting,
		"bills":                bills,
		"batch":                batch,
		"bank":                 bank,
	})
}

func FetchVoucherTypeHandler(c *gin.Context){
	
}