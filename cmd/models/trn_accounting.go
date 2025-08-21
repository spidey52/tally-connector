package models

type TrnAccounting struct {
	Guid   string  `db:"guid" json:"guid"`
	Ledger string  `db:"ledger" json:"ledger"`
	Amount float64 `db:"amount" json:"amount"`
}
