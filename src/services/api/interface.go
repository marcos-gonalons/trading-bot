package api

import (
	"TradingBot/src/types"
	"time"
)

type Trade struct {
	Side         string
	Size         float64
	InitialPrice float64
	FinalPrice   float64
	Result       float64
	OpenedAt     time.Time
	ClosedAt     time.Time
}

// Interface implemented by the broker APIs
type Interface interface {
	Login() (*AccessToken, error)
	GetQuote(marketName string) (*Quote, error)
	CreateOrder(order *Order) error
	GetOrders() ([]*Order, error)
	SetOrders(orders []*Order)
	GetWorkingOrders(orders []*Order) []*Order
	ModifyOrder(order *Order) error
	CloseOrder(orderID string) error
	SetPositions(positions []*Position)
	GetPositions() ([]*Position, error)
	ClosePosition(marketName string) error
	AddTrade(
		order *Order,
		position *Position,
		slippageFunc func(price float64, order *Order) float64,
		eurExchangeRate float64,
		lastCandle *types.Candle,
		marketData *types.MarketData,
	)
	GetTrades() []*Trade
	SetTrades([]*Trade)
	CloseAllOrders() error
	CloseAllPositions() error
	GetState() (*State, error)
	SetState(state *State)
	ModifyPosition(marketName string, takeProfit *string, stopLoss *string) error

	SetTimeout(t time.Duration)
	GetTimeout() time.Duration

	IsSessionDisconnectedError(err error) bool
	IsOrderAlreadyExistsError(err error) bool
	IsPositionAlreadyExistsError(err error) bool
	IsNotEnoughFundsError(err error) bool
	IsOrderPendingCancelError(err error) bool
	IsOrderCancelledError(err error) bool
	IsOrderFilledError(err error) bool
	IsInvalidHoursError(err error) bool
	IsClosePositionRequestInProgressError(err error) bool
	IsPositionNotFoundError(err error) bool

	IsWorkingOrder(order *Order) bool
	IsLimitOrder(order *Order) bool
	IsStopOrder(order *Order) bool
	IsMarketOrder(order *Order) bool
	IsLongOrder(order *Order) bool
	IsShortOrder(order *Order) bool
	IsLongPosition(position *Position) bool
	IsShortPosition(position *Position) bool

	GetBracketOrders(marketName string) (slOrder, tpOrder *Order)
	GetWorkingOrderWithBracketOrders(side string, marketName string, orders []*Order) []*Order
	GetSLAndTPOrders(parentID string, orders []*Order) (*Order, *Order)
}

type DataInterface interface {
	SetOrders(orders []*Order)
	GetOrders() []*Order
	SetPositions(positions []*Position)
	GetPositions() []*Position
	SetState(state *State)
	GetState() *State
}
