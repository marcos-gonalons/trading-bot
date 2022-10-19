package interfaces

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/logger"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/types"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// todo: move this method to a dedicated class
type OnValidTradeSetupParams struct {
	Price              float64
	StopLossDistance   float32
	TakeProfitDistance float32
	RiskPercentage     float64
	IsValidTime        bool
	Side               string
	WithPendingOrders  bool
	OrderType          string
	MinPositionSize    int64
}

type MarketInstanceDependencies struct {
	APIRetryFacade          retryFacade.Interface
	API                     api.Interface
	APIData                 api.DataInterface
	Logger                  logger.Interface
	CandlesHandler          candlesHandler.Interface
	HorizontalLevelsService horizontalLevels.Interface
	TrendsService           trends.Interface
}

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
	GetEurExchangeRate() float64
	SetCurrentPositionExecutedAt(timestamp int64)
	GetAPIData() api.DataInterface

	Log(message string)
	SavePendingOrder(side string, validTimes *types.TradingTimes)
	GetPendingOrder() *api.Order
	CreatePendingOrder(side string)
	SetPendingOrder(order *api.Order)
	CheckIfSLShouldBeAdjusted(params *types.MarketStrategyParams, position *api.Position)
	CheckOpenPositionTTL(params *types.MarketStrategyParams, position *api.Position)
	OnValidTradeSetup(params OnValidTradeSetupParams)

	GetCandlesHandler() candlesHandler.Interface

	SetDependencies(MarketInstanceDependencies)
}
