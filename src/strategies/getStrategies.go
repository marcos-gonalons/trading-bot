package strategies

import (
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/strategies/tickers/AUDUSD"
	"TradingBot/src/strategies/tickers/EURUSD"
	"TradingBot/src/strategies/tickers/GBPUSD"
	"TradingBot/src/strategies/tickers/NZDUSD"
	"TradingBot/src/strategies/tickers/USDCAD"
	"TradingBot/src/strategies/tickers/USDCHF"
	"TradingBot/src/strategies/tickers/USDJPY"
	"TradingBot/src/strategies/tickers/interfaces"
	//"TradingBot/src/strategies/tickers/GER30"
)

func (s *Handler) getStrategies() []interfaces.TickerInterface {
	return []interfaces.TickerInterface{
		//s.getStrategy(GER30.GetStrategyInstance(s.API, s.APIRetryFacade, s.Logger)),
		s.getStrategy(EURUSD.GetStrategyInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getStrategy(GBPUSD.GetStrategyInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getStrategy(USDCAD.GetStrategyInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getStrategy(USDJPY.GetStrategyInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getStrategy(USDCHF.GetStrategyInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getStrategy(NZDUSD.GetStrategyInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getStrategy(AUDUSD.GetStrategyInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
	}
}

func (s *Handler) getStrategy(strategy interfaces.TickerInterface) interfaces.TickerInterface {
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
