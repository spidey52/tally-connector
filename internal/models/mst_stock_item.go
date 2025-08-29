package models

import "github.com/jackc/pgx/v5/pgtype"

type MstStockItem struct {
	Guid              string         `db:"guid" json:"guid"`
	Alterid           int32          `db:"alterid" json:"alterid"`
	Name              string         `db:"name" json:"name"`
	Parent            string         `db:"parent" json:"parent"`
	Alias             string         `db:"alias" json:"alias"`
	Description       string         `db:"description" json:"description"`
	Notes             string         `db:"notes" json:"notes"`
	PartNumber        string         `db:"part_number" json:"part_number"`
	Uom               string         `db:"uom" json:"uom"`
	AlternateUom      string         `db:"alternate_uom" json:"alternate_uom"`
	Conversion        int32          `db:"conversion" json:"conversion"`
	OpeningBalance    pgtype.Numeric `db:"opening_balance" json:"opening_balance"`
	OpeningRate       pgtype.Numeric `db:"opening_rate" json:"opening_rate"`
	OpeningValue      pgtype.Numeric `db:"opening_value" json:"opening_value"`
	ClosingBalance    pgtype.Numeric `db:"closing_balance" json:"closing_balance"`
	ClosingRate       pgtype.Numeric `db:"closing_rate" json:"closing_rate"`
	ClosingValue      pgtype.Numeric `db:"closing_value" json:"closing_value"`
	CostingMethod     string         `db:"costing_method" json:"costing_method"`
	GstTypeOfSupply   *string        `db:"gst_type_of_supply" json:"gst_type_of_supply"`
	GstHsnCode        *string        `db:"gst_hsn_code" json:"gst_hsn_code"`
	GstHsnDescription *string        `db:"gst_hsn_description" json:"gst_hsn_description"`
	GstRate           *int32         `db:"gst_rate" json:"gst_rate"`
	GstTaxability     *string        `db:"gst_taxability" json:"gst_taxability"`
}
