package models

/*
	guid string
	ledger string
	amount  float64
	additional_alloacation_type string
*/

type TrnInventoryAccounting struct {
	Guid                     string  `json:"guid" db:"guid"`
	Ledger                   string  `json:"ledger" db:"ledger"`
	Amount                   float64 `json:"amount" db:"amount"`
	AdditionalAllocationType string  `json:"additional_allocation_type" db:"additional_allocation_type"`
}
