package handler

import (
	"tally-connector/internal/db"
	"tally-connector/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func FetchStockItem(c *gin.Context) {
	var ctx = c.Request.Context()
	var stockItems []models.MstStockItem

	q := psql.Select(
		sm.Columns("guid", "name", "parent", "alias", "description", "notes"),
		sm.From("mst_stock_item"),
	)

	query, args, err := q.Build(c)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to build query",
		})
		return
	}

	err = pgxscan.Select(ctx, db.GetDB(), &stockItems, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve stock items",
		})
		return
	}

	c.JSON(200, gin.H{"stock_items": stockItems})
}
