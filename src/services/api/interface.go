package api

// Interface implemented by the broker APIs
type Interface interface {
	Login(username, password string) (*AccessToken, error)
	GetQuote(symbol string) (*Quote, error)
	CreateOrder(order *Order) error
	GetOrders() ([]*Order, error)
	ModifyOrder(order *Order) error
	CloseOrder(orderID int64) error
	GetPositions() ([]*Position, error)
	ClosePosition() error
}
