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
		s.setupMarketInstance(EURUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		/*s.getMarket(GBPUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(USDCAD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(USDCHF.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(NZDUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),
		s.getMarket(AUDUSD.GetMarketInstance(s.API, s.APIData, s.APIRetryFacade, s.Logger)),*/
	}
}

func (s *Handler) setupMarketInstance(market interfaces.MarketInterface) interfaces.MarketInterface {
	candlesHandler := &candlesHandler.Service{
		Logger:            s.Logger,
		MarketData:        market.GetMarketData(),
		IndicatorsBuilder: indicators.GetInstance(),
	}

	market.SetCandlesHandler(candlesHandler)
	market.SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))
	market.SetTrendsService(trends.GetServiceInstance())

	return market
}
