package interfaces

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/types"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// todo: move this method to a dedicated class
type OnValidTradeSetupParams struct {
	Price              float64
	StrategyName       string
	StopLossDistance   float32
	TakeProfitDistance float32
	RiskPercentage     float64
	IsValidTime        bool
	Side               string
	WithPendingOrders  bool
	OrderType          string
	MinPositionSize    int64
}

// MarketInterface ...
// todo: sort the methods and add a description
// and maybe divide the interface into smaller ones.
type MarketInterface interface {
	Initialize()
	DailyReset()
	OnReceiveMarketData(data *tradingviewsocket.QuoteData)
	OnNewCandle()
	GetSocketMarketName() string
	GetAPIMarketName() string
	GetCurrentBrokerQuote() *api.Quote
	GetMarketType() types.MarketType
	IsTradeableOnWeekends() bool
	GetTradingHours() *types.TradingHours
	SetCurrentBrokerQuote(quote *api.Quote)
	GetTimeframe() *types.Timeframe
	GetMarketData() *types.MarketData
	SetEurExchangeRate(rate float64)
	GetEurExchangeRate() float64
	SetCurrentPositionExecutedAt(timestamp int64)
	GetAPIData() api.DataInterface

	Log(strategyName string, message string)
	SavePendingOrder(side string, validTimes *types.TradingTimes)
	GetPendingOrder() *api.Order
	CreatePendingOrder(side string)
	SetPendingOrder(order *api.Order)
	CheckIfSLShouldBeAdjusted(params *types.MarketStrategyParams, position *api.Position)
	CheckOpenPositionTTL(params *types.MarketStrategyParams, position *api.Position)
	OnValidTradeSetup(params OnValidTradeSetupParams)

	SetCandlesHandler(candlesHandler candlesHandler.Interface)
	GetCandlesHandler() candlesHandler.Interface
	SetHorizontalLevelsService(horizontalLevelsService horizontalLevels.Interface)
	GetHorizontalLevelsService() horizontalLevels.Interface
	SetTrendsService(trendsService trends.Interface)
	GetTrendsService() trends.Interface
}
