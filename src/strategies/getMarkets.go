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
	instances := []interfaces.MarketInterface{
		EURUSD.GetMarketInstance(),
		// GBPUSD.GetMarketInstance(),
		// etc etc
	}

	for _, instance := range instances {
		candlesHandler := &candlesHandler.Service{
			Logger:            s.Logger,
			MarketData:        instance.GetMarketData(),
			IndicatorsBuilder: indicators.GetInstance(),
		}

		// todo: maybe move the trends and horizontallevels services to the handler and get them here via s.*****
		dependencies := interfaces.MarketInstanceDependencies{
			APIRetryFacade:          s.APIRetryFacade,
			API:                     s.API,
			APIData:                 s.APIData,
			Logger:                  s.Logger,
			CandlesHandler:          candlesHandler,
			TrendsService:           trends.GetServiceInstance(),
			HorizontalLevelsService: horizontalLevels.GetServiceInstance(candlesHandler),
		}

		instance.SetDependencies(dependencies)
	}

	return instances
}
