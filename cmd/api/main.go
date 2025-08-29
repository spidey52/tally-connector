package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"tally-connector/cmd/api/handler"
	"tally-connector/cmd/api/middlewares"
	"tally-connector/internal/db"
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

	syncTableGroup := r.Group("/sync-tables")
	syncTableGroup.GET("", handler.GetSyncTables)

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
