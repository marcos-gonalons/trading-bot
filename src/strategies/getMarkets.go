package strategies

import (
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/candlesHandler/indicators"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/strategies/markets/EURUSD"
	"TradingBot/src/strategies/markets/interfaces"
	//"TradingBot/src/strategies/markets/GER30"
)

// GetMarkets...
func (s *Handler) GetMarkets() []interfaces.MarketInterface {
	return []interfaces.MarketInterface{
		//s.getMarket(GER30.GetMarketInstance(s.API, s.APIRetryFacade, s.Logger)),
		s.getMarket(EURUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		/*s.getMarket(GBPUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(USDCAD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(USDJPY.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(USDCHF.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(NZDUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(AUDUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),*/
	}
}

func (s *Handler) getMarket(strategy interfaces.MarketInterface) interfaces.MarketInterface {
	candlesHandler := &candlesHandler.Service{
		Logger:            s.Logger,
		Symbol:            strategy.Parent().GetSymbol(),
		Timeframe:         *strategy.Parent().GetTimeframe(),
		IndicatorsBuilder: indicators.GetInstance(),
	}

	strategy.Parent().SetCandlesHandler(candlesHandler)
	strategy.Parent().SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))
	strategy.Parent().SetTrendsService(trends.GetServiceInstance())

	return strategy
}
