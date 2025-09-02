package handler

import (
	"context"
	"sort"
	"tally-connector/api/middlewares"
	"tally-connector/internal/db"
	"tally-connector/internal/models"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func getLastSyncTime(ctx context.Context) (map[string]time.Time, error) {
	result := make(map[string]time.Time)

	var dbresults []models.SyncLog

	err := pgxscan.Select(ctx, db.GetDB(), &dbresults, "SELECT table_name, max(end_time) AS end_time FROM tbl_sync_logs GROUP BY table_name ORDER BY end_time;")

	if err != nil {
		return nil, err
	}

	for _, dbresult := range dbresults {
		result[dbresult.Table] = dbresult.EndTime
	}

	return result, nil
}

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

	if err != nil {
		c.JSON(500, gin.H{
			"status":  "failed to fetch sync tables",
			"message": err.Error(),
		})
		return
	}

	lastSyncTime, err := getLastSyncTime(ctx)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "failed to fetch last sync time",
			"message": err.Error(),
		})
		return
	}

	for idx, table := range syncTables {
		if lst, ok := lastSyncTime[table.Name]; ok {
			syncTables[idx].LastSyncTime = lst
		}
	}

	sort.Slice(syncTables, func(j, i int) bool {
		return syncTables[i].LastSyncTime.Before(syncTables[j].LastSyncTime)
	})

	c.JSON(200, gin.H{
		"count":  len(syncTables),
		"result": syncTables,
	})
}

type SyncLogQueryParams struct {
	middlewares.PaginationParams
	Status string `form:"status"`
}

func GetSyncLogs(c *gin.Context) {
	var ctx = c.Request.Context()

	var queryParams SyncLogQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(400, gin.H{
			"status":  "failed to bind query params",
			"message": err.Error(),
		})
		return
	}

	q := psql.Select(
		sm.Columns("table_name", "group_name", "start_time", "end_time", "duration", "status", "message"),
		sm.From("tbl_sync_logs"),
		sm.OrderBy("start_time DESC"),
		sm.Limit(queryParams.Limit),
		sm.Offset(queryParams.Page*queryParams.Limit),
	)

	if queryParams.Status != "" {
		q.Apply(sm.Where(psql.Quote("status").EQ(psql.Arg(queryParams.Status))))
	}

	if queryParams.Search != "" {
		q.Apply(sm.Where(psql.Quote("table_name").ILike(psql.Arg("%" + queryParams.Search + "%"))))
	}

	if queryParams.StartDate != "" && queryParams.EndDate != "" {
		q.Apply(sm.Where(psql.Quote("start_time").GTE(psql.Arg(queryParams.StartDate))))
		q.Apply(sm.Where(psql.Quote("end_time").LTE(psql.Arg(queryParams.EndDate))))
	}

	query, args, err := q.Build(ctx)

	if err != nil {
		c.JSON(500, gin.H{
			"status":  "failed to build query",
			"message": err.Error(),
		})
		return
	}

	var syncLogs = []models.SyncLog{}

	err = pgxscan.Select(ctx, db.GetDB(), &syncLogs, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"status":  "failed to fetch sync logs",
			"message": err.Error(),
		})
		return
	}

	var totalCount int64
	countQuery := psql.Select(
		sm.Columns("COUNT(*)"),
		sm.From("tbl_sync_logs"),
	)

	if queryParams.Status != "" {
		countQuery.Apply(sm.Where(psql.Quote("status").EQ(psql.Arg(queryParams.Status))))
	}

	if queryParams.StartDate != "" && queryParams.EndDate != "" {
		countQuery.Apply(sm.Where(psql.Quote("start_time").GTE(psql.Arg(queryParams.StartDate))))
		countQuery.Apply(sm.Where(psql.Quote("end_time").LTE(psql.Arg(queryParams.EndDate))))
	}

	query, args, err = countQuery.Build(ctx)

	if err != nil {
		c.JSON(500, gin.H{
			"status":  "failed to build count query",
			"message": err.Error(),
		})
		return
	}

	err = pgxscan.Get(ctx, db.GetDB(), &totalCount, query, args...)

	if err != nil {
		c.JSON(500, gin.H{
			"status":  "failed to fetch sync logs count",
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"count":  len(syncLogs),
		"result": syncLogs,
	})
}

func SeedSyncTables(c *gin.Context) {
	ctx := c.Request.Context()
	models.CreateSyncTables()
	db.SeedSyncTables(ctx)
	c.JSON(201, gin.H{
		"status":  "success",
		"message": "sync table created successfully",
	})
}
