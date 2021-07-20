package strategies

import (
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/strategies/interfaces"
	"TradingBot/src/strategies/tickers/GER30"
)

func (s *Handler) getStrategies() []interfaces.StrategyInterface {
	return []interfaces.StrategyInterface{
		s.getGER30Strategy(),
	}
}

func (s *Handler) getGER30Strategy() interfaces.StrategyInterface {
	GER30Strategy := GER30.GetStrategyInstance(s.API, s.APIRetryFacade, s.Logger)
	candlesHandler := &candlesHandler.Service{
		Logger:    s.Logger,
		Symbol:    GER30Strategy.Parent().GetSymbol().BrokerAPIName,
		Timeframe: *GER30Strategy.Parent().GetTimeframe(),
	}
	GER30Strategy.Parent().SetCandlesHandler(candlesHandler)
	GER30Strategy.Parent().SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))
	GER30Strategy.Parent().SetTrendsService(trends.GetServiceInstance())

	return GER30Strategy
}
