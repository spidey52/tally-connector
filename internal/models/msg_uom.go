package models

type MstUom struct {
	Guid            string `db:"guid" json:"guid"`
	Name            string `db:"name" json:"name"`
	FormalName      string `db:"formalname" json:"formalname"`
	IsSimpleUnit    int    `db:"is_simple_unit" json:"is_simple_unit"`
	BaseUnits       string `db:"base_units" json:"base_units"`
	AdditionalUnits string `db:"additional_units" json:"additional_units"`
	Conversion      int    `db:"conversion" json:"conversion"`
}
