package models

type MstStockItem struct {
	Guid        string `db:"guid" json:"guid"`
	Name        string `db:"name" json:"name"`
	Quantity    int    `db:"quantity" json:"quantity"`
	UnitPrice   int    `db:"unit_price" json:"unit_price"`
	TotalPrice  int    `db:"total_price" json:"total_price"`
	Category    string `db:"category" json:"category"`
	SubCategory string `db:"sub_category" json:"sub_category"`
}

