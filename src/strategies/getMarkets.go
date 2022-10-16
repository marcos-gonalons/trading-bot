package strategies

import (
	"TradingBot/src/markets/EURUSD"
	"TradingBot/src/markets/interfaces"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/candlesHandler/indicators"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
)

// GetMarkets...
func (s *Handler) GetMarkets() []interfaces.MarketInterface {
	return []interfaces.MarketInterface{
		s.getMarket(EURUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		/*s.getMarket(GBPUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(USDCAD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(USDCHF.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(NZDUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(AUDUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),*/
	}
}

func (s *Handler) getMarket(strategy interfaces.MarketInterface) interfaces.MarketInterface {
	candlesHandler := &candlesHandler.Service{
		Logger:            s.Logger,
		Market:            strategy.Parent().GetMarket(),
		Timeframe:         *strategy.Parent().GetTimeframe(),
		IndicatorsBuilder: indicators.GetInstance(),
	}

	strategy.Parent().SetCandlesHandler(candlesHandler)
	strategy.Parent().SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))
	strategy.Parent().SetTrendsService(trends.GetServiceInstance())

	return strategy
}
