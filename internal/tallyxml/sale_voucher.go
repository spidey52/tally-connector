package tallyxml

import "encoding/xml"

type SaleVoucher struct {
	XMLName          xml.Name             `xml:"VOUCHER"`
	Action           string               `xml:"ACTION,attr"`
	Date             string               `xml:"DATE"`
	VoucherNumber    string               `xml:"VOUCHERNUMBER"`
	PartyLedgerName  string               `xml:"PARTYLEDGERNAME"`
	VoucherTypeName  string               `xml:"VOUCHERTYPENAME"`
	PersistedView    string               `xml:"PERSISTEDVIEW"`
	IsInvoice        string               `xml:"ISINVOICE"`
	VchEntryMode     string               `xml:"VCHENTRYMODE"`
	Narration        string               `xml:"NARRATION,omitempty"`
	LedgerEntries    []SaleLedgerEntry    `xml:"LEDGERENTRIES.LIST"`
	InventoryEntries []SaleInventoryEntry `xml:"ALLINVENTORYENTRIES.LIST"`
}

type SaleLedgerEntry struct {
	LedgerName string `xml:"LEDGERNAME"`
	IsPositive string `xml:"ISDEEMEDPOSITIVE"`
	Amount     string `xml:"AMOUNT"`
}

type SaleInventoryEntry struct {
	StockItemName string                `xml:"STOCKITEMNAME"`
	Rate          string                `xml:"RATE"`
	ActualQty     string                `xml:"ACTUALQTY"`
	BilledQty     string                `xml:"BILLEDQTY"`
	IsPositive    string                `xml:"ISDEEMEDPOSITIVE"`
	Amount        string                `xml:"AMOUNT"`
	Allocations   []SaleAccountingAlloc `xml:"ACCOUNTINGALLOCATIONS.LIST"`
}

type SaleAccountingAlloc struct {
	LedgerName string `xml:"LEDGERNAME"`
	Amount     string `xml:"AMOUNT"`
	IsPositive string `xml:"ISDEEMEDPOSITIVE"`
}
