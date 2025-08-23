package models

type MstVoucherType struct {
	Guid             string `db:"guid" json:"guid"`
	AlterID          int64  `db:"alterid" json:"alterid"`
	Name             string `db:"name" json:"name"`
	Parent           string `db:"parent" json:"parent"`
	NumberingMethod  string `db:"numbering_method" json:"numbering_method"`
	IsDeemedPositive int    `db:"is_deemedpositive" json:"is_deemedpositive"`
	AffectsStock     int    `db:"affects_stock" json:"affects_stock"`
}
