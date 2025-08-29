package models

import "time"

type TrnVoucher struct {
	Guid                string     `json:"guid" db:"guid"`
	AlterID             int64      `json:"alterid" db:"alterid"`
	Date                time.Time  `json:"date" db:"date"`
	VoucherType         string     `json:"voucher_type" db:"voucher_type"`
	VoucherNumber       string     `json:"voucher_number" db:"voucher_number"`
	ReferenceNumber     string     `json:"reference_number" db:"reference_number"`
	ReferenceDate       *time.Time `json:"reference_date" db:"reference_date"`
	Narration           string     `json:"narration" db:"narration"`
	PartyName           string     `json:"party_name" db:"party_name"`
	PlaceOfSupply       string     `json:"place_of_supply" db:"place_of_supply"`
	IsInvoice           *int       `json:"is_invoice" db:"is_invoice"`
	IsAccountingVoucher *int       `json:"is_accounting_voucher" db:"is_accounting_voucher"`
	IsInventoryVoucher  *int       `json:"is_inventory_voucher" db:"is_inventory_voucher"`
	IsOrderVoucher      *int       `json:"is_order_voucher" db:"is_order_voucher"`
}
