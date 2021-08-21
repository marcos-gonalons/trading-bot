package strategies

import (
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/strategies/interfaces"
	"TradingBot/src/strategies/tickers/EURUSD"
	"TradingBot/src/strategies/tickers/GBPUSD"
	"TradingBot/src/strategies/tickers/GER30"
)

func (s *Handler) getStrategies() []interfaces.StrategyInterface {
	return []interfaces.StrategyInterface{
		s.getStrategy(GER30.GetStrategyInstance(s.API, s.APIRetryFacade, s.Logger)),
		s.getStrategy(EURUSD.GetStrategyInstance(s.API, s.APIRetryFacade, s.Logger)),
		s.getStrategy(GBPUSD.GetStrategyInstance(s.API, s.APIRetryFacade, s.Logger)),
	}
}

func (s *Handler) getStrategy(strategy interfaces.StrategyInterface) interfaces.StrategyInterface {
	candlesHandler := &candlesHandler.Service{
		Logger:    s.Logger,
		Symbol:    strategy.Parent().GetSymbol(),
		Timeframe: *strategy.Parent().GetTimeframe(),
	}
	strategy.Parent().SetCandlesHandler(candlesHandler)
	strategy.Parent().SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))
	strategy.Parent().SetTrendsService(trends.GetServiceInstance())

	return strategy
}
