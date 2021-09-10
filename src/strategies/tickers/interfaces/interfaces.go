package interfaces

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/types"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// TickerInterface ...
type TickerInterface interface {
	Parent() BaseTickerClassInterface
	Initialize()
	DailyReset()
	OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData)
}

// BaseTickerClassInterface ...
type BaseTickerClassInterface interface {
	Initialize()
	DailyReset()
	SetCandlesHandler(candlesHandler candlesHandler.Interface)
	SetHorizontalLevelsService(horizontalLevelsService horizontalLevels.Interface)
	SetTrendsService(trendsService trends.Interface)
	OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData)
	SetOrders(orders []*api.Order)
	GetOrders() []*api.Order
	SetCurrentBrokerQuote(quote *api.Quote)
	GetCurrentBrokerQuote() *api.Quote
	SetPositions(positions []*api.Position)
	GetPositions() []*api.Position
	SetState(state *api.State)
	GetState() *api.State
	GetTimeframe() *types.Timeframe
	GetSymbol() *types.Symbol
}
