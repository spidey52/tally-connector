package models

type TrnBill struct {
	Guid             string `db:"guid" json:"guid"`
	Ledger           string `db:"ledger" json:"ledger"`
	Name             string `db:"name" json:"name"`
	Amount           string `db:"amount" json:"amount"`
	BillType         string `db:"bill_type" json:"billtype"`
	BillCreditPeriod int    `db:"bill_credit_period" json:"bill_credit_period"`
}
