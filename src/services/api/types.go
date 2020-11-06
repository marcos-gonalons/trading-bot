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
	Ask    float32
	Bid    float32
	Price  float32
	Volume float64
}

// Order ...
type Order struct {
	ID           int64
	CurrentAsk   float32
	CurrentBid   float32
	DurationType string
	Instrument   string
	Qty          float32
	Side         string
	StopLoss     float32
	TakeProfit   float32
	Type         string
	LimitPrice   float32
	StopPrice    float32
	Status       string
	ParentID     int64
}

// Position ...
type Position struct {
	Instrument   string
	Qty          float32
	Side         string
	AvgPrice     float32
	UnrealizedPl float32
}
