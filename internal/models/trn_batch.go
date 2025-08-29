package models

type TrnBatch struct {
	Guid     string  `json:"guid" db:"guid"`
	Item     string  `json:"item" db:"item"`
	Name     string  `json:"name" db:"name"`
	Quantity float64 `json:"quantity" db:"quantity"`
	Amount   float64 `json:"amount" db:"amount"`
	Godown   string  `json:"godown" db:"godown"`
}
