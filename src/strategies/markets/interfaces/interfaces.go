package interfaces

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/types"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// MarketInterface ...
type MarketInterface interface {
	Parent() BaseMarketClassInterface
	Initialize()
	DailyReset()
	OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData)
	OnNewCandle()
}

// BaseMarketClassInterface ...
type BaseMarketClassInterface interface {
	SetCandlesHandler(candlesHandler candlesHandler.Interface)
	GetCandlesHandler() candlesHandler.Interface
	SetHorizontalLevelsService(horizontalLevelsService horizontalLevels.Interface)
	SetTrendsService(trendsService trends.Interface)
	OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData)
	OnNewCandle()
	SetCurrentBrokerQuote(quote *api.Quote)
	GetCurrentBrokerQuote() *api.Quote
	GetTimeframe() *types.Timeframe
	GetSymbol() *types.Symbol
	SetEurExchangeRate(rate float64)
	GetEurExchangeRate() float64
}
