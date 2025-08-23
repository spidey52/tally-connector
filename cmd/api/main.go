package main

import (
	"flag"
	"fmt"
	"tally-connector/cmd/api/handler"
	"tally-connector/cmd/api/middlewares"
	"tally-connector/cmd/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db.ConnectDB("postgres://satyam:satyam52@localhost:5432/tally_db")

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.Use(middlewares.InsertPaginationParams)

	r.GET("/fetch-ledgers", handler.FetchLedgerHandler)
	r.GET("/fetch-ledgers-autocomplete", handler.FetchLedgerAutoComplete)

	r.GET("/fetch-vouchers", handler.FetchVoucherHandler)
	r.GET("/fetch-vouchers/:id", handler.FetchVoucherDetailsHandler)

	r.GET("/fetch-voucher-type", handler.FetchVoucherTypeHandler)

	r.POST("/sales-voucher", handler.CreateSalesVoucher)

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
