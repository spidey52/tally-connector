package models

import "time"

type TrnBank struct {
	Guid             string     `json:"guid" db:"guid"`
	Ledger           string     `json:"ledger" db:"ledger"`
	TransactionType  string     `json:"transaction_type" db:"transaction_type"`
	InstrumentDate   *time.Time `json:"instrument_date" db:"instrument_date"`
	InstrumentNumber string     `json:"instrument_number" db:"instrument_number"`
	BankName         string     `json:"bank_name" db:"bank_name"`
	Amount           float64    `json:"amount" db:"amount"`
	BankersDate      *time.Time `json:"bankers_date" db:"bankers_date"`
}
