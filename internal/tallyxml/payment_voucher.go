package tallyxml

import "encoding/xml"

type PaymentVoucher struct {
	TallyVoucher
	LedgerEntries []BankLedgerEntry `xml:"ALLLEDGERENTRIES.LIST"`
}

type BankLedgerEntry struct {
	LedgerEntry
	Allocations []BankAllocations `xml:"BANKALLOCATIONS.LIST"`
}

type BankAllocations struct {
	XMLName         xml.Name `xml:"BANKALLOCATIONS"`
	Date            string   `xml:"DATE"`
	InstrumentDate  string   `xml:"INSTRUMENTDATE"`
	TransactionType string   `xml:"TRANSACTIONTYPE"`
	BankPartyName   string   `xml:"BANKPARTYNAME"`
	Amount          float64  `xml:"AMOUNT"`
}
