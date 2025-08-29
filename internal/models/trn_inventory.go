package models

/*
guid string
item string
quantity float
rate float
amount flaot
additional_amount number
discount float
godown string
*/
type TrnInventory struct {
	Item             string
	Quantity         int
	Rate             float64
	Amount           float64
	AdditionalAmount float64
	DiscountAmount   float64
}
