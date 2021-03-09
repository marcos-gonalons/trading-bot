package api

import "time"

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
	CloseAllOrders() error
	CloseAllPositions() error
	GetState() (*State, error)
	ModifyPosition(symbol string, takeProfit *string, stopLoss *string) error
	SetTimeout(t time.Duration)
	GetTimeout() time.Duration
	IsSessionDisconnectedError(err error) bool
}
