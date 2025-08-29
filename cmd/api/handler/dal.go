package handler

import (
	"math"
	"tally-connector/cmd/api/middlewares"
	"tally-connector/internal/db"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

type Dal struct {
	Guid          string    `json:"guid" db:"guid"`
	Date          time.Time `json:"date" db:"date"`
	VoucherNumber string    `json:"voucher_number" db:"voucher_number"`
	VoucherType   string    `json:"voucher_type" db:"voucher_type"`
	Ledger        string    `json:"ledger" db:"ledger"`
	Amount        float64   `json:"amount" db:"amount"`

	// computed value
	DebitAmount  float64 `json:"debit_amount" db:"-"`
	CreditAmount float64 `json:"credit_amount" db:"-"`
}

func FetchDal(c *gin.Context) {
	ctx := c.Request.Context()
	var ledgerID, _ = c.GetQuery("ledger_id")
	var voucherType, _ = c.GetQuery("voucher_type")

	queryParams := c.MustGet("pagination_params").(middlewares.PaginationParams)

	if ledgerID == "" {
		c.JSON(400, gin.H{
			"error":   "Bad Request",
			"message": "ledger_id is required",
		})
		return
	}

	if queryParams.StartDate == "" || queryParams.EndDate == "" {
		c.JSON(400, gin.H{
			"error":   "Bad Request",
			"message": "start_date and end_date are required",
		})
		return
	}

	inner := psql.Select(
		sm.Columns("DISTINCT a.guid"),
		sm.From("trn_accounting a"),
		sm.InnerJoin("trn_voucher tv_in").
			On(psql.Quote("tv_in", "guid").EQ(psql.Quote("a", "guid"))),
		sm.Where(psql.Quote("a", "ledger").EQ(psql.Arg(ledgerID))),
		sm.Where(psql.Quote("tv_in", "date").GTE(psql.Arg(queryParams.StartDate)).And(psql.Quote("tv_in", "date").LT(psql.Arg(queryParams.EndDate)))),
	)

	if voucherType != "" {
		inner.Apply(
			sm.Where(psql.Quote("tv_in", "voucher_type").EQ(psql.Arg(voucherType))),
		)
	}

	sub := psql.Select(
		sm.Columns(
			"ta.guid",
			"ta.ledger",
			"ta.amount",
			"ROW_NUMBER() OVER (PARTITION BY ta.guid ORDER BY ta.amount DESC) AS rn",
		),
		sm.From("trn_accounting").As("ta"),
		sm.Where(psql.Quote("ta", "ledger").NE(psql.Arg(ledgerID))),
		sm.Where(psql.Quote("ta", "guid").In(inner)),
	)

	// ⬇️ alias the subquery here, in `From`
	q := psql.Select(
		sm.Columns("x.guid", "tv.date", "tv.voucher_number", "tv.voucher_type", "x.ledger", "x.amount"),
		sm.From(sub).As("x"), // ✅ alias applied here
		sm.InnerJoin("trn_voucher tv").
			On(psql.Quote("tv", "guid").EQ(psql.Quote("x", "guid"))),
		sm.Where(psql.Quote("x", "rn").EQ(psql.Arg(1))),
		sm.Limit(queryParams.Limit),
		sm.Offset(queryParams.Limit*queryParams.Page),
	)

	query, args, err := q.Build(ctx)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	var dal = []Dal{}

	err = pgxscan.Select(ctx, db.GetDB(), &dal, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	q = psql.Select(
		sm.Columns("COUNT(*)"),
		sm.From(sub).As("x"),
		sm.InnerJoin("trn_voucher tv").
			On(psql.Quote("tv", "guid").EQ(psql.Quote("x", "guid"))),
		sm.Where(psql.Quote("x", "rn").EQ(psql.Arg(1))),
	)

	query, args, err = q.Build(ctx)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	var count int64
	err = pgxscan.Get(ctx, db.GetDB(), &count, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	var balance, _ = getTotalBalance(ctx, []string{ledgerID}, queryParams.StartDate, queryParams.EndDate)

	for i := range dal {
		isPositive := dal[i].Amount >= 0

		absAmount := math.Abs(dal[i].Amount)

		// categorize the amount into debit and credit
		// for daybook it should be opposite.
		if !isPositive {
			dal[i].CreditAmount = absAmount
		} else {
			dal[i].DebitAmount = absAmount
		}

	}
	c.JSON(200, gin.H{
		"count":   count,
		"data":    dal,
		"balance": balance[ledgerID],
	})
}
