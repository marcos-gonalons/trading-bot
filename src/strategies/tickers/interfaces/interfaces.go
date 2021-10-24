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
	SetCandlesHandler(candlesHandler candlesHandler.Interface)
	SetHorizontalLevelsService(horizontalLevelsService horizontalLevels.Interface)
	SetTrendsService(trendsService trends.Interface)
	OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData)
	SetCurrentBrokerQuote(quote *api.Quote)
	GetCurrentBrokerQuote() *api.Quote
	GetTimeframe() *types.Timeframe
	GetSymbol() *types.Symbol
}
