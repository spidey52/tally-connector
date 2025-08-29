package handler

import (
	"fmt"
	"tally-connector/internal/db"
	"tally-connector/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func GetSyncTables(c *gin.Context) {
	var ctx = c.Request.Context()

	q := psql.Select(
		sm.Columns("table_name", "group_name", "max_wait", "sync_interval"),
		sm.From("tbl_sync_tables"),
	)

	query, args, err := q.Build(ctx)

	if err != nil {
		c.JSON(500, gin.H{
			"status":  "failed to build query",
			"message": err.Error(),
		})
		return
	}

	var syncTables = []models.SyncTable{}

	err = pgxscan.Select(ctx, db.GetDB(), &syncTables, query, args...)

	fmt.Println("Query:", query)
	fmt.Println("Args:", args)

	if err != nil {
		c.JSON(500, gin.H{
			"status":  "failed to fetch sync tables",
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"count":  len(syncTables),
		"result": syncTables,
	})
}
