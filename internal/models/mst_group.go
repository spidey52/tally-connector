package models

type MstGroup struct {
	Guid               string `json:"guid" db:"guid"`
	Name               string `json:"name" db:"name"`
	Parent             string `json:"parent" db:"parent"`
	PrimaryGroup       string `json:"primary_group" db:"primary_group"`
	IsRevenue          int    `json:"is_revenue" db:"is_revenue"`
	IsDeemedPositive   int    `json:"is_deemedpositive" db:"is_deemedpositive"`
	IsReserved         int    `json:"is_reserved" db:"is_reserved"`
	AffectsGrossProfit int    `json:"affects_gross_profit" db:"affects_gross_profit"`
	SortPosition       int    `json:"sort_position" db:"sort_position"`
}
