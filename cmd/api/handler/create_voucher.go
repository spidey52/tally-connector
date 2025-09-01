package handler

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"tally-connector/internal/db"
	"tally-connector/internal/helper"
	"tally-connector/internal/models"
	"tally-connector/internal/tallyxml"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
)

// ---------------------- DTOs (input from frontend) ----------------------

type TallyProducts struct {
	Name     string  `json:"name" binding:"required"`
	Rate     float64 `json:"rate" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required"`
}

type TallyDiscounts struct {
	Type   string  `json:"type" binding:"required,oneof=SAUDA DISCOUNT LOADING DISCOUNT"`
	Amount float64 `json:"amount" binding:"required"`
}

type VoucherDto struct {
	Date          string `json:"date" binding:"required,datetime=20060102"`
	PartyName     string `json:"party_name" binding:"required"`
	VoucherNumber string `json:"voucher_number" binding:"required"`
	VoucherType   string `json:"voucher_type" binding:"required"`
	Narration     string `json:"narration"`
}

type CreateSalesVoucherDto struct {
	VoucherDto
	Discounts []TallyDiscounts `json:"discounts"`
	Products  []TallyProducts  `json:"products" binding:"required"`
}

// ---------------------- Globals ----------------------

var client = &http.Client{}
var tallyEndpoint = "http://172.16.0.47:9000"

// ---------------------- Helper functions ----------------------

func buildSaleVoucher(dto CreateSalesVoucherDto) tallyxml.SaleVoucher {
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
			ActualQty:     fmt.Sprintf("%.2f", p.Quantity),
			BilledQty:     fmt.Sprintf("%.2f", p.Quantity),
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
		Narration:        dto.Narration,
	}
}

func callTallyApi(ctx context.Context, env tallyxml.Envelope) error {
	xmlData, err := xml.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}

	log.Println("---- Request XML ----")
	log.Println(string(xmlData))

	// if 120%5 == 0 {
	// 	return nil
	// }

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

	if tallyResp.LineError != "" {
		return fmt.Errorf("tally error: %s", tallyResp.LineError)
	}

	if tallyResp.Errors > 0 {
		return fmt.Errorf("tally error: %d errors found", tallyResp.Errors)
	}
	if tallyResp.Exceptions > 0 {
		return fmt.Errorf("tally error: %d exceptions found", tallyResp.Exceptions)
	}

	return nil
}

func buildPaymentVoucher(dto CreatePaymentVoucherDto) tallyxml.PaymentVoucher {
	return tallyxml.PaymentVoucher{
		TallyVoucher: tallyxml.TallyVoucher{
			Action:          "Create",
			Date:            dto.Date,
			VoucherNumber:   dto.VoucherNumber,
			PartyLedgerName: dto.PartyName,
			VoucherTypeName: dto.VoucherType,
			Narration:       dto.Narration,
		},
		LedgerEntries: []tallyxml.BankLedgerEntry{
			{
				LedgerEntry: tallyxml.LedgerEntry{
					LedgerName: dto.PartyName,
					IsPositive: "Yes",
					Amount:     fmt.Sprintf("%.2f", -dto.Amount),
				},
			},
			{
				LedgerEntry: tallyxml.LedgerEntry{
					LedgerName: dto.TargetBank,
					IsPositive: "No",
					Amount:     fmt.Sprintf("%.2f", dto.Amount),
				},
				Allocations: []tallyxml.BankAllocations{
					{
						Date:            dto.Date,
						InstrumentDate:  dto.Date,
						TransactionType: "Cheque",
						BankPartyName:   dto.PartyName,
						Amount:          dto.Amount,
					},
				},
			},
		},
	}
}

// ---------------------- Handlers ----------------------

func CreateSalesVoucher(c *gin.Context) {
	ctx := c.Request.Context()

	var dto []CreateSalesVoucherDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// ledgerMap := ledger

	ledgerMap := helper.LedgerMapByAlias()

	var vouchers []tallyxml.SaleVoucher
	for idx, d := range dto {
		// dto[idx].PartyName = ledgerMap[d.PartyName]
		dto[idx].PartyName = ledgerMap[d.PartyName]
		vouchers = append(vouchers, buildSaleVoucher(d))
	}

	env := tallyxml.Envelope{
		Header: tallyxml.Header{TallyRequest: "Import Data"},
		Body: tallyxml.Body{
			ImportData: tallyxml.ImportData{
				RequestDesc: tallyxml.RequestDesc{ReportName: "Vouchers"},
				RequestData: tallyxml.RequestData{
					TallyMessage: tallyxml.TallyMessage{Voucher: &vouchers},
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

type CreatePaymentVoucherDto struct {
	VoucherDto
	Amount     float64 `json:"amount" binding:"required"`
	TargetBank string  `json:"target_bank" binding:"required"`
}

func CreatePaymentVoucher(c *gin.Context) {
	ctx := c.Request.Context()

	var dto []CreatePaymentVoucherDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var paymentVouchers []tallyxml.PaymentVoucher
	for _, d := range dto {
		paymentVouchers = append(paymentVouchers, buildPaymentVoucher(d))
	}

	env := tallyxml.Envelope{
		Header: tallyxml.Header{TallyRequest: "Import Data"},
		Body: tallyxml.Body{
			ImportData: tallyxml.ImportData{
				RequestDesc: tallyxml.RequestDesc{ReportName: "Vouchers"},
				RequestData: tallyxml.RequestData{
					TallyMessage: tallyxml.TallyMessage{
						Voucher: paymentVouchers,
					},
				},
			},
		},
	}

	if err := callTallyApi(ctx, env); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "payment vouchers created successfully"})
}

func invalidLedgers(ledgers []string) []string {
	results := make([]string, 0)

	// Query using Postgres ANY() to handle string slices
	stmt := `
		SELECT name, alias
		FROM mst_ledger
		WHERE name = ANY($1) OR alias = ANY($1)
	`

	var dbledgers []models.Ledger

	err := pgxscan.Select(
		context.Background(),
		db.GetDB(),
		&dbledgers,
		stmt,
		ledgers, // pass the slice directly
	)
	if err != nil {
		log.Println("Error in fetching ledgers:", err)
		return results
	}

	// Build a lookup map of available ledgers
	availableLedgers := map[string]bool{}
	for _, l := range dbledgers {
		if l.Name != "" {
			availableLedgers[l.Name] = true
		}
		if l.Alias != "" {
			availableLedgers[l.Alias] = true
		}
	}

	// Collect invalid ones
	for _, l := range ledgers {
		if _, ok := availableLedgers[l]; !ok {
			results = append(results, l)
		}
	}

	return results
}

func existingVouchers(vouchers []string) []string {
	var voucherNumbers []string

	stmt := `
		SELECT voucher_number
		FROM trn_voucher
		WHERE voucher_number = ANY($1)
	`

	err := pgxscan.Select(
		context.Background(),
		db.GetDB(),
		&voucherNumbers,
		stmt,
		vouchers, // pass the slice directly
	)

	if err != nil {
		log.Println("Error in fetching existing vouchers:", err)
		return []string{}
	}

	return voucherNumbers
}

type ValidateVoucherDto struct {
	Ledgers  []string `json:"ledgers"`
	Vouchers []string `json:"vouchers"`
}

func ValidateVouchers(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.JSON(405, gin.H{"error": "Method not allowed"})
		return
	}

	var dto ValidateVoucherDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ledgers := invalidLedgers(dto.Ledgers)
	vouchers := existingVouchers(dto.Vouchers)

	c.JSON(200, gin.H{
		"invalid_ledgers":   ledgers,
		"existing_vouchers": vouchers,
	})

}
