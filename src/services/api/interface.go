package api

import "time"

// Interface implemented by the broker APIs
type Interface interface {
	Login() (*AccessToken, error)
	GetQuote(symbol string) (*Quote, error)
	CreateOrder(order *Order) error
	GetOrders() ([]*Order, error)
	GetWorkingOrders(orders []*Order) []*Order
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
	IsOrderAlreadyExistsError(err error) bool
	IsNotEnoughFundsError(err error) bool
	IsOrderPendingCancelError(err error) bool
	IsOrderCancelledError(err error) bool
	IsOrderFilledError(err error) bool
	IsInvalidHoursError(err error) bool
	IsClosePositionRequestInProgressError(err error) bool

	IsWorkingOrder(order *Order) bool
	IsLimitOrder(order *Order) bool
	IsStopOrder(order *Order) bool
	IsLongOrder(order *Order) bool
	IsShortOrder(order *Order) bool
	IsLongPosition(position *Position) bool
	IsShortPosition(position *Position) bool

	GetBracketOrdersForOpenedPosition(position *Position) (slOrder, tpOrder *Order)
	GetWorkingOrderWithBracketOrders(side string, symbol string, orders []*Order) []*Order
}
