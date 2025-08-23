package handler

import (
	"tally-connector/cmd/db"
	"tally-connector/cmd/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func FetchVoucherTypeHandler(c *gin.Context) {
	var ctx = c.Request.Context()

	var voucherTypes []models.MstVoucherType
	q := psql.Select(
		sm.Columns("guid", "name", "parent", "is_deemedpositive", "affects_stock", "numbering_method"),
		sm.From("mst_vouchertype"),
	)

	query, args, err := q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Select(ctx, db.GetDB(), &voucherTypes, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve voucher types",
		})
		return
	}

	c.JSON(200, gin.H{
		"count":         len(voucherTypes),
		"voucher_types": voucherTypes,
	})
}
