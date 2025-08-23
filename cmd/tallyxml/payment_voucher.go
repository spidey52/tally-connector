package tallyxml

import "encoding/xml"

type PaymentVoucher struct {
	Action        string  `xml:"ACTION,attr"`
	PartyName     string  `xml:"PARTY_NAME"`
	VoucherNumber string  `xml:"VOUCHER_NUMBER"`
	Date          string  `xml:"DATE"`
	Amount        float64 `xml:"AMOUNT"`
	Remarks       string  `xml:"NARRATION"`
}

type PaymentLedgerEntry struct {
	XMLName     xml.Name `xml:"LEDGERENTRY"`
	LedgerName  string   `xml:"LEDGERNAME"`
	IsDeemedPos string   `xml:"ISDEEMEDPOSITIVE"`
	Amount      float64  `xml:"AMOUNT"`
}

// <VOUCHER ACTION="Create">
//     <DATE>20250807</DATE>
//     <NARRATION>testing by satyam </NARRATION>
//     <VOUCHERNUMBER>1010</VOUCHERNUMBER>
//     <PARTYLEDGERNAME>DPAT9327</PARTYLEDGERNAME>
//     <VOUCHERTYPENAME>IRTGS</VOUCHERTYPENAME>
//      <ALLLEDGERENTRIES.LIST>
//         <LEDGERNAME>DPAT9327</LEDGERNAME>
//         <ISDEEMEDPOSITIVE>No</ISDEEMEDPOSITIVE>
//         <AMOUNT>12000</AMOUNT>
//     </ALLLEDGERENTRIES.LIST>
//     <ALLLEDGERENTRIES.LIST>
//         <LEDGERNAME>MIND 7071 </LEDGERNAME>
//         <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
//         <AMOUNT>-12000</AMOUNT>
//         <BANKALLOCATIONS.LIST>
//             <DATE>20250802</DATE>
//             <INSTRUMENTDATE>20250807</INSTRUMENTDATE>
//             <TRANSACTIONTYPE>Cheque/DD</TRANSACTIONTYPE>
//             <BANKPARTYNAME>DPAT9327</BANKPARTYNAME>
//             <AMOUNT>-12000.00</AMOUNT>
//         </BANKALLOCATIONS.LIST>
//     </ALLLEDGERENTRIES.LIST>

// </VOUCHER>
