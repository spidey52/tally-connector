package models

type Ledger struct {
	Name           string  `db:"name" json:"name"`
	Parent         string  `db:"parent" json:"parent"`
	Alias          string  `db:"alias" json:"alias"`
	OpeningBalance float64 `db:"opening_balance" json:"opening_balance"`
	ClosingBalance float64 `db:"closing_balance" json:"closing_balance"`

	// calculated field... coming from the sum of amount(trn_accounting)
	TotalNetDebit  float64 `db:"total_net_debit" json:"total_net_debit"`
	TotalNetCredit float64 `db:"total_net_credit" json:"total_net_credit"`
}
