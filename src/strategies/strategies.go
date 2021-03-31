package strategies

import (
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/strategies/breakoutAnticipation"
	"TradingBot/src/strategies/test"
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
		Symbol:    breakoutAnticipationStrategy.GetSymbolForAPI(),
		Timeframe: *breakoutAnticipationStrategy.GetTimeframe(),
	}
	breakoutAnticipationStrategy.SetCandlesHandler(candlesHandler)
	breakoutAnticipationStrategy.SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))

	return breakoutAnticipationStrategy
}

func (s *Handler) getTestStrategy() Interface {
	testStrategy := test.GetStrategyInstance(
		s.API,
		s.APIRetryFacade,
		s.Logger,
	)
	candlesHandler := &candlesHandler.Service{
		Logger:    s.Logger,
		Symbol:    testStrategy.GetSymbolForAPI(),
		Timeframe: *testStrategy.GetTimeframe(),
	}
	testStrategy.SetCandlesHandler(candlesHandler)
	testStrategy.SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))

	return testStrategy
}
