package handler

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"tally-connector/internal/tallyxml"

	"github.com/gin-gonic/gin"
)

// ---------------------- DTOs (input from frontend) ----------------------

type TallyProducts struct {
	Name     string  `json:"name" binding:"required"`
	Rate     float64 `json:"rate" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Quantity int     `json:"quantity" binding:"required"`
}

type TallyDiscounts struct {
	Type   string  `json:"type" binding:"required,oneof=SAUDA DISCOUNT LOADING DISCOUNT"`
	Amount float64 `json:"amount" binding:"required"`
}

type CreateSalesVoucherDto struct {
	PartyName     string           `json:"party_name" binding:"required"`
	VoucherNumber string           `json:"voucher_number" binding:"required"`
	Date          string           `json:"date" binding:"required,datetime=20060102"`
	Discounts     []TallyDiscounts `json:"discounts"`
	Products      []TallyProducts  `json:"products" binding:"required"`
}

// ---------------------- Globals ----------------------

var client = &http.Client{}
var tallyEndpoint = "http://100.77.107.9:9000"

// ---------------------- Helper functions ----------------------

func buildVoucher(dto CreateSalesVoucherDto) tallyxml.SaleVoucher {
	var totalProductValue float64
	var ledgerEntries []tallyxml.SaleLedgerEntry
	var inventoryEntries []tallyxml.SaleInventoryEntry

	// Products
	for _, p := range dto.Products {
		totalProductValue += p.Amount

		inventoryEntries = append(inventoryEntries, tallyxml.SaleInventoryEntry{
			StockItemName: p.Name,
			Rate:          fmt.Sprintf("%.2f", p.Rate),
			Amount:        fmt.Sprintf("%.2f", p.Amount),
			ActualQty:     fmt.Sprintf("%d", p.Quantity),
			BilledQty:     fmt.Sprintf("%d", p.Quantity),
			IsPositive:    "No",
			Allocations: []tallyxml.SaleAccountingAlloc{
				{
					LedgerName: "SALE",
					IsPositive: "No",
					Amount:     fmt.Sprintf("%.2f", p.Amount),
				},
			},
		})
	}

	// Discounts
	var totalDiscount float64
	for _, d := range dto.Discounts {
		totalDiscount += d.Amount
		ledgerEntries = append(ledgerEntries, tallyxml.SaleLedgerEntry{
			LedgerName: d.Type,
			IsPositive: "No",
			Amount:     fmt.Sprintf("%.2f", -d.Amount),
		})
	}

	// Party ledger entry
	ledgerEntries = append([]tallyxml.SaleLedgerEntry{
		{
			LedgerName: dto.PartyName,
			IsPositive: "Yes",
			Amount:     fmt.Sprintf("%.2f", -(totalProductValue - totalDiscount)),
		},
	}, ledgerEntries...)

	return tallyxml.SaleVoucher{
		Action:           "Create",
		Date:             dto.Date,
		VoucherNumber:    dto.VoucherNumber,
		PartyLedgerName:  dto.PartyName,
		VoucherTypeName:  "SALE IMPORT",
		PersistedView:    "Invoice Voucher View",
		IsInvoice:        "Yes",
		VchEntryMode:     "Item Invoice",
		LedgerEntries:    ledgerEntries,
		InventoryEntries: inventoryEntries,
	}
}

func callTallyApi(ctx context.Context, env tallyxml.Envelope) error {
	xmlData, err := xml.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	log.Println("---- Request XML ----")
	log.Println(string(xmlData))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tallyEndpoint, bytes.NewReader(xmlData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/xml")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Tally API response status: %s", resp.Status)
		return fmt.Errorf("failed to call Tally API: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	log.Println("---- Tally Response ----")
	log.Println(string(body))

	var tallyResp tallyxml.TallyResponse
	if err := xml.Unmarshal(body, &tallyResp); err != nil {
		return err
	}

	log.Printf("Tally Response Parsed: %+v\n", tallyResp)

	if tallyResp.Created == 0 {
		return fmt.Errorf("tally did not create the voucher, response: %+v", tallyResp)
	}

	return nil
}

// ---------------------- Handlers ----------------------

func CreateSalesVoucher(c *gin.Context) {
	ctx := c.Request.Context()

	var dto CreateSalesVoucherDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	voucher := buildVoucher(dto)

	env := tallyxml.Envelope{
		Header: tallyxml.Header{TallyRequest: "Import Data"},
		Body: tallyxml.Body{
			ImportData: tallyxml.ImportData{
				RequestDesc: tallyxml.RequestDesc{ReportName: "Vouchers"},
				RequestData: tallyxml.RequestData{
					TallyMessage: tallyxml.TallyMessage{Voucher: &voucher},
				},
			},
		},
	}

	if err := callTallyApi(ctx, env); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Sales voucher created successfully"})
}
