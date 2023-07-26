package emaCrossover

import (
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/strategies"
)

func OnNewCandle(market markets.MarketInterface) {
	if market.GetMarketData().EmaCrossoverSetup == nil {
		return
	}

	if market.GetMarketData().EmaCrossoverSetup.LongSetupParams != nil {
		market.Log("Calling EmaCrossoverLongs strategy")
		EmaCrossoverLongs(strategies.Params{
			Type:                 ibroker.LongSide,
			MarketStrategyParams: market.GetMarketData().EmaCrossoverSetup.LongSetupParams,
			MarketData:           market.GetMarketData(),
			CandlesHandler:       market.GetCandlesHandler(),
			Market:               market,
		})
	}

	if market.GetMarketData().EmaCrossoverSetup.ShortSetupParams != nil {
		market.Log("Calling EmaCrossoverShorts strategy")
		EmaCrossoverShorts(strategies.Params{
			Type:                 ibroker.ShortSide,
			MarketStrategyParams: market.GetMarketData().EmaCrossoverSetup.ShortSetupParams,
			MarketData:           market.GetMarketData(),
			CandlesHandler:       market.GetCandlesHandler(),
			Market:               market,
		})
	}
}
