package tradingviewsocket

// SocketMessage ...
type SocketMessage struct {
	Message string      `json:"m"`
	Payload interface{} `json:"p"`
}

// QuoteMessage ...
type QuoteMessage struct {
	Symbol string     `mapstructure:"n"`
	Status string     `mapstructure:"s"`
	Data   *QuoteData `mapstructure:"v"`
}

// QuoteData ...
type QuoteData struct {
	Price  *float64 `mapstructure:"lp"`
	Volume *float64 `mapstructure:"volume"`
	Bid    *float64 `mapstructure:"bid"`
	Ask    *float64 `mapstructure:"ask"`
}
