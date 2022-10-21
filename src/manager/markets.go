package manager

import (
	"TradingBot/src/markets"
	"TradingBot/src/markets/EURUSD"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/candlesHandler/indicators"
)

func (s *Manager) GetMarkets() []markets.MarketInterface {
	instances := []markets.MarketInterface{
		EURUSD.GetMarketInstance(s.ServicesContainer),
		// GBPUSD.GetMarketInstance(s.ServicesContainer),
		// etc etc
	}

	for _, instance := range instances {
		instance.SetCandlesHandler(&candlesHandler.Service{
			Logger:            s.ServicesContainer.Logger,
			MarketData:        instance.GetMarketData(),
			IndicatorsBuilder: indicators.GetInstance(),
		})
	}

	return instances
}
