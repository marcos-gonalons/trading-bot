package manager

import (
	"TradingBot/src/markets"
	"TradingBot/src/markets/AUDUSD"
	"TradingBot/src/markets/DAX"
	"TradingBot/src/markets/EUROSTOXX"
	"TradingBot/src/markets/EURUSD"
	"TradingBot/src/markets/GBPUSD"
	"TradingBot/src/markets/NZDUSD"
	"TradingBot/src/markets/SP500"
	"TradingBot/src/markets/USDCAD"
	"TradingBot/src/markets/USDCHF"
	"TradingBot/src/services/candlesHandler"
)

func (s *Manager) GetMarkets() []markets.MarketInterface {
	instances := []markets.MarketInterface{
		AUDUSD.GetMarketInstance(),
		EURUSD.GetMarketInstance(),
		GBPUSD.GetMarketInstance(),
		NZDUSD.GetMarketInstance(),
		USDCAD.GetMarketInstance(),
		USDCHF.GetMarketInstance(),

		DAX.GetMarketInstance(),
		EUROSTOXX.GetMarketInstance(),
		SP500.GetMarketInstance(),
	}

	for _, instance := range instances {
		instance.SetContainer(s.ServicesContainer)
		instance.SetCandlesHandler(&candlesHandler.Service{
			Logger:            s.ServicesContainer.Logger,
			MarketData:        instance.GetMarketData(),
			IndicatorsBuilder: s.ServicesContainer.IndicatorsService,
		})
	}

	return instances
}
