package markets

import (
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/types"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// MarketInterface ...
// todo: sort the methods and add a description
// and maybe divide the interface into smaller ones.
type MarketInterface interface {
	Initialize()
	DailyReset()
	OnReceiveMarketData(data *tradingviewsocket.QuoteData)
	OnNewCandle()
	GetCurrentBrokerQuote() *api.Quote
	SetCurrentBrokerQuote(quote *api.Quote)
	GetMarketData() *types.MarketData
	SetCurrentPositionExecutedAt(timestamp int64)

	Log(message string)
	SavePendingOrder(side string, validTimes *types.TradingTimes)
	GetPendingOrder() *api.Order
	CreatePendingOrder(side string)
	SetPendingOrder(order *api.Order)
	CheckIfSLShouldBeAdjusted(params *types.MarketStrategyParams, position *api.Position)
	CheckOpenPositionTTL(params *types.MarketStrategyParams, position *api.Position)
	OnValidTradeSetup(params OnValidTradeSetupParams)

	SetContainer(*services.Container)
	SetCandlesHandler(candlesHandler.Interface)
	GetCandlesHandler() candlesHandler.Interface
}
