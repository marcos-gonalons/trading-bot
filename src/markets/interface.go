package markets

import (
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/types"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

type MarketInterface interface {
	Initialize()
	DailyReset()
	OnReceiveMarketData(data *tradingviewsocket.QuoteData)
	OnNewCandle()
	GetMarketData() *types.MarketData
	SetCurrentPositionExecutedAt(timestamp int64)

	Log(message string)
	SavePendingOrder(side string, validTimes *types.TradingTimes)
	GetPendingOrder() *api.Order
	CreatePendingOrder(side string)
	SetPendingOrder(order *api.Order)
	CheckOpenPositionTTL(params *types.MarketStrategyParams, position *api.Position)
	OnValidTradeSetup(params OnValidTradeSetupParams)

	SetStrategyParams(longs *types.MarketStrategyParams, shorts *types.MarketStrategyParams)

	SetContainer(*services.Container)
	SetCandlesHandler(candlesHandler.Interface)
	GetCandlesHandler() candlesHandler.Interface
}
