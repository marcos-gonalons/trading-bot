package retryFacade

import (
	"TradingBot/src/services/api"
)

// Interface implemented by the api retry facade
type Interface interface {
	Login(retryParams RetryParams)
	CloseOrders(orders []*api.Order, retryParams RetryParams)
	ClosePositions(retryParams RetryParams)
	ClosePosition(marketName string, retryParams RetryParams)
	ModifyPosition(marketName string, tp string, sl string, retryParams RetryParams)
	CreateOrder(order *api.Order, setStringValues func(order *api.Order), retryParams RetryParams)
}
