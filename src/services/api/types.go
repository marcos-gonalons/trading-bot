package api

import "time"

// Credentials ...
type Credentials struct {
	Username  string
	Password  string
	AccountID string
}

// AccessToken ...
type AccessToken struct {
	Token      string
	Expiration time.Time
}

// Quote ...
type Quote struct {
	Ask    float64
	Bid    float64
	Price  float64
	Volume float64
}

// Order ...
type Order struct {
	ID           string
	CurrentAsk   *float64
	CurrentBid   *float64
	DurationType string
	Instrument   string
	Qty          float64
	Side         string
	StopLoss     *float64
	TakeProfit   *float64
	Type         string
	LimitPrice   *float64
	StopPrice    *float64
	Status       string
	ParentID     *string
	StringValues *OrderStringValues
}

// OrderStringValues ...
type OrderStringValues struct {
	CurrentAsk *string
	CurrentBid *string
	Qty        *string
	StopLoss   *string
	TakeProfit *string
	LimitPrice *string
	StopPrice  *string
}

// Position ...
type Position struct {
	Instrument   string
	Qty          float64
	Side         string
	AvgPrice     float64
	UnrealizedPl float64
	CreatedAt    *int64
}

// State ...
type State struct {
	Balance      float64
	UnrealizedPL float64
	Equity       float64
}
