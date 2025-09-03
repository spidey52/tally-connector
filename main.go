package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"tally-connector/api/handler"
	"tally-connector/api/middlewares"
	"tally-connector/internal/db"
	"tally-connector/internal/jobs"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type LoaderEnv struct {
	PostgresURL   string `env:"POSTGRES_URL"`
	TallyEndpoint string `env:"TALLY_ENDPOINT"`
}

var env LoaderEnv

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	env.PostgresURL = os.Getenv("POSTGRES_URL")
	env.TallyEndpoint = os.Getenv("TALLY_ENDPOINT")
}

func main() {
	db.ConnectDB(env.PostgresURL)

	var ctx = context.Background()
	go jobs.ProcessImportQueue(ctx)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	}))

	r.Use(middlewares.InsertPaginationParams)

	r.GET("/fetch-ledgers", handler.FetchLedgerHandler)
	r.GET("/fetch-groups", handler.FetchGroupHandler)
	r.POST("/create-ledger", handler.CreateLedgerHandler)
	r.GET("/fetch-ledgers-autocomplete", handler.FetchLedgerAutoComplete)

	r.GET("/fetch-vouchers", handler.FetchVoucherHandler)
	r.GET("/fetch-daybook", handler.FetchDaybook)
	// r.GET("/fetch-dal", handler.FetchDal)
	r.POST("/fetch-dal", handler.FetchDal)
	r.GET("/fetch-products", handler.FetchStockItem)
	r.GET("/fetch-stock-items", handler.FetchStockItem)
	r.GET("/fetch-vouchers/:id", handler.FetchVoucherDetailsHandler)

	r.GET("/fetch-voucher-type", handler.FetchVoucherTypeHandler)

	r.POST("/sales-voucher", handler.CreateSalesVoucher)
	r.POST("/payment-voucher", handler.CreatePaymentVoucher)
	r.POST("/validate-voucher", handler.ValidateVouchers)

	r.POST("/sync", func(c *gin.Context) {
		type SyncDto struct {
			Filters []string `json:"filters" oneof:"sync,async"`
			Mode    string   `json:"mode"`
		}

		var dto SyncDto
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if len(dto.Filters) == 0 {
			c.JSON(400, gin.H{"error": "no filters provided", "message": "please provide at least one filter"})
			return
		}

		for _, filter := range dto.Filters {
			jobs.GetDefaultWorkerPool().AddJob(&jobs.Job{
				ID:     filter,
				Status: "queued",
			})
		}

		c.JSON(200, gin.H{
			"message": "added in import queue",
		})
	})

	r.GET("/sync-jobs", func(c *gin.Context) {
		jobsList := jobs.GetDefaultWorkerPool().GetJobs()

		var inqueue int
		var processing int

		for _, job := range jobsList {
			switch job.Status {
			case "queued":
				inqueue++
			case "processing":
				processing++
			}
		}

		c.JSON(200, gin.H{
			"inqueue":    inqueue,
			"processing": processing,
			"jobs":       jobsList,
		})
	})

	loaderRoutes := r.Group("/loaderRoutes")
	loaderRoutes.GET("/tables", handler.GetSyncTables)
	loaderRoutes.GET("logs", handler.GetSyncLogs)
	loaderRoutes.POST("/seed", handler.SeedSyncTables)

	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	r.Run(fmt.Sprintf(":%d", *port))

}

/*
	required model to create
	stock item
	stock group

	// opening bill allocation
	// opening batch allocation

	mst ledger
	mst rate
	mst group
	mst godown
	mst cost category

*/
