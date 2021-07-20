package strategies

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/types"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// Interface implemented by all the strategies
type Interface interface {
	Initialize()
	DailyReset()
	SetCandlesHandler(candlesHandler candlesHandler.Interface)
	SetHorizontalLevelsService(horizontalLevelsService horizontalLevels.Interface)
	SetTrendsService(trendsService trends.Interface)
	OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData)
	SetOrders(orders []*api.Order)
	SetCurrentBrokerQuote(quote *api.Quote)
	SetPositions(positions []*api.Position)
	SetState(state *api.State)
	GetTimeframe() *types.Timeframe
	GetSymbol() *types.Symbol
}
