package tradingviewsocket

// MarketData ...
type MarketData struct {
	Ask    float64
	Bid    float64
	Price  float64
	Volume float64
}

// SocketMessage ...
type SocketMessage struct {
	Message string      `json:"m"`
	Payload interface{} `json:"p"`
}
