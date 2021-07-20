package strategies

import (
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/strategies/GER30"
	"TradingBot/src/strategies/testStrategy"
)

func (s *Handler) getStrategies() []Interface {
	return []Interface{
		s.getGER30Strategy(),
		// s.getTestStrategy(),
	}
}

func (s *Handler) getGER30Strategy() Interface {
	GER30Strategy := GER30.GetStrategyInstance(
		s.API,
		s.APIRetryFacade,
		s.Logger,
	)
	candlesHandler := &candlesHandler.Service{
		Logger:    s.Logger,
		Symbol:    GER30Strategy.GetSymbol().BrokerAPIName,
		Timeframe: *GER30Strategy.GetTimeframe(),
	}
	GER30Strategy.SetCandlesHandler(candlesHandler)
	GER30Strategy.SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))
	GER30Strategy.SetTrendsService(trends.GetServiceInstance())

	return GER30Strategy
}

func (s *Handler) getTestStrategy() Interface {
	testStrategy := testStrategy.GetStrategyInstance(
		s.API,
		s.APIRetryFacade,
		s.Logger,
	)
	candlesHandler := &candlesHandler.Service{
		Logger:    s.Logger,
		Symbol:    testStrategy.GetSymbol().BrokerAPIName,
		Timeframe: *testStrategy.GetTimeframe(),
	}
	testStrategy.SetCandlesHandler(candlesHandler)
	testStrategy.SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))
	testStrategy.SetTrendsService(trends.GetServiceInstance())

	return testStrategy
}
