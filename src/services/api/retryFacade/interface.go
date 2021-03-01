package retryFacade

import (
	"TradingBot/src/services/api"
)

// Interface implemented by the api retry facade
type Interface interface {
	Login(retryParams RetryParams)
	CloseSpecificOrders(orders []*api.Order, retryParams RetryParams)
	CloseAllWorkingOrders(retryParams RetryParams)
	ClosePositions(retryParams RetryParams)
	ModifyPosition(symbol string, tp string, sl string, retryParams RetryParams)
	CreateOrder(order *api.Order, getCurrentBrokerQuote func() *api.Quote, setStringValues func(order *api.Order), retryParams RetryParams)
}
