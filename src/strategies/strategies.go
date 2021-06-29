package strategies

import (
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/strategies/breakoutAnticipation"
	"TradingBot/src/strategies/testStrategy"
)

func (s *Handler) getStrategies() []Interface {
	return []Interface{
		s.getBreakoutAnticipationStrategy(),
		// s.getTestStrategy(),
	}
}

func (s *Handler) getBreakoutAnticipationStrategy() Interface {
	breakoutAnticipationStrategy := breakoutAnticipation.GetStrategyInstance(
		s.API,
		s.APIRetryFacade,
		s.Logger,
	)
	candlesHandler := &candlesHandler.Service{
		Logger:    s.Logger,
		Symbol:    breakoutAnticipationStrategy.GetSymbol().BrokerAPIName,
		Timeframe: *breakoutAnticipationStrategy.GetTimeframe(),
	}
	breakoutAnticipationStrategy.SetCandlesHandler(candlesHandler)
	breakoutAnticipationStrategy.SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))

	return breakoutAnticipationStrategy
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

	return testStrategy
}
