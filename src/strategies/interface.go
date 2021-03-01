package strategies

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/candlesHandler"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper"
)

// Interface implemented by all the strategies
type Interface interface {
	Initialize()
	Reset()
	SetCandlesHandler(candlesHandler candlesHandler.Interface)
	OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData)
	SetOrders(orders []*api.Order)
	SetCurrentBrokerQuote(quote *api.Quote)
	SetPositions(positions []*api.Position)
	SetState(state *api.State)
}
