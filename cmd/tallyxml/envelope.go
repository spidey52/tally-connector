package tallyxml

import "encoding/xml"

type Envelope struct {
	XMLName xml.Name `xml:"ENVELOPE"`
	Header  Header   `xml:"HEADER"`
	Body    Body     `xml:"BODY"`
}

type Header struct {
	TallyRequest string `xml:"TALLYREQUEST"`
}

type Body struct {
	ImportData ImportData `xml:"IMPORTDATA"`
}

type ImportData struct {
	RequestDesc RequestDesc `xml:"REQUESTDESC"`
	RequestData RequestData `xml:"REQUESTDATA"`
}

type RequestDesc struct {
	ReportName string `xml:"REPORTNAME"`
}

type RequestData struct {
	XMLName      xml.Name `xml:"REQUESTDATA"`
	TallyMessage any      `xml:"TALLYMESSAGE"`
}

type Voucher struct {
	XMLName         xml.Name `xml:"VOUCHER"`
	Action          string   `xml:"ACTION,attr"`
	Date            string   `xml:"DATE"`
	VoucherNumber   string   `xml:"VOUCHERNUMBER"`
	PartyLedgerName string   `xml:"PARTYLEDGERNAME"`
	VoucherTypeName string   `xml:"VOUCHERTYPENAME"`
	Narration       string   `xml:"NARRATION,omitempty"`
}

type LedgerEntry struct {
	XMLName    xml.Name `xml:"LEDGER"`
	LedgerName string   `xml:"LEDGERNAME"`
	IsPositive string   `xml:"ISDEEMEDPOSITIVE"`
	Amount     string   `xml:"AMOUNT"`
}

// This is where we allow flexibility
type TallyMessage struct {
	Voucher *SaleVoucher `xml:"VOUCHER,omitempty"`
	// Ledger    *Ledger    `xml:"LEDGER,omitempty"`
	// StockItem *StockItem `xml:"STOCKITEM,omitempty"`
	// add more types later if needed
}

type TallyResponse struct {
	Created    int `xml:"CREATED"`
	Altered    int `xml:"ALTERED"`
	Deleted    int `xml:"DELETED"`
	LastVchId  int `xml:"LASTVCHID"`
	LastMid    int `xml:"LASTMID"`
	Combined   int `xml:"COMBINED"`
	Ignored    int `xml:"IGNORED"`
	Errors     int `xml:"ERRORS"`
	Cancelled  int `xml:"CANCELLED"`
	Exceptions int `xml:"EXCEPTIONS"`
}
