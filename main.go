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
	"tally-connector/internal/loader"
	"tally-connector/internal/redisclient"
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
	go jobs.ProcessImportQueue()

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
			IsAll   bool     `json:"is_all"`
			Mode    string   `json:"mode"`
		}

		var dto SyncDto
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if dto.Mode == "sync" || dto.Mode == "" {
			loader.ImportAll(dto.Filters...)

			c.JSON(200, gin.H{
				"message": "sync completed",
			})
			return
		}

		if len(dto.Filters) == 0 {
			c.JSON(400, gin.H{"error": "no filters provided", "message": "please provide at least one filter"})
			return
		}

		redisclient.GetRedisClient().LPush(context.Background(), redisclient.ImportQueueKey, dto.Filters)

		c.JSON(200, gin.H{
			"message": "added in import queue",
		})
	})

	loader := r.Group("/loader")
	loader.GET("/tables", handler.GetSyncTables)
	loader.GET("logs", handler.GetSyncLogs)
	loader.POST("/seed", handler.SeedSyncTables)

	// loader.POST("", handler.CreateSyncTable)

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
