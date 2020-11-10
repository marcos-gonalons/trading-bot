package api

// Interface implemented by the broker APIs
type Interface interface {
	Login() (*AccessToken, error)
	GetQuote(symbol string) (*Quote, error)
	CreateOrder(order *Order) error
	GetOrders() ([]*Order, error)
	ModifyOrder(order *Order) error
	CloseOrder(orderID string) error
	GetPositions() ([]*Position, error)
	ClosePosition(symbol string) error
	CloseEverything() error
	GetState() (*State, error)
	ModifyPosition(symbol string, takeProfit *float32, stopLoss *float32) error
}
