package strategies

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/logger"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/strategies/interfaces"
	"TradingBot/src/strategies/tickers/GER30"
)

func (s *Handler) getStrategies(
	API api.Interface,
	APIRetryFacade retryFacade.Interface,
	logger logger.Interface,
) []interfaces.StrategyInterface {
	return []interfaces.StrategyInterface{
		s.getGER30Strategy(API, APIRetryFacade, logger),
	}
}

func (s *Handler) getGER30Strategy(
	API api.Interface,
	APIRetryFacade retryFacade.Interface,
	logger logger.Interface,
) interfaces.StrategyInterface {
	GER30Strategy := GER30.GetStrategyInstance(API, APIRetryFacade, logger)
	candlesHandler := &candlesHandler.Service{
		Logger:    logger,
		Symbol:    GER30Strategy.Parent().GetSymbol().BrokerAPIName,
		Timeframe: *GER30Strategy.Parent().GetTimeframe(),
	}
	GER30Strategy.Parent().SetCandlesHandler(candlesHandler)
	GER30Strategy.Parent().SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))
	GER30Strategy.Parent().SetTrendsService(trends.GetServiceInstance())

	return GER30Strategy
}
