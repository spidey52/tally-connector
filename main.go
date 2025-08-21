package main

import (
	"tally-connector/cmd/api/handler"
	"tally-connector/cmd/db"

	"github.com/gin-gonic/gin"
)

func main() {
	db.ConnectDB()

	r := gin.Default()

	r.GET("/fetch-ledgers", handler.FetchLedgerHandler)
	r.GET("/fetch-vouchers", handler.FetchVoucherHandler)
	r.GET("/fetch-vouchers/:id", handler.FetchVoucherDetailsHandler)

	r.Run(":8080")

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
