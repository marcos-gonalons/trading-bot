package emaCrossover

import (
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/strategies"
)

var NAME = "emaCrossover"

func OnNewCandle(market markets.MarketInterface) {
	if market.GetMarketData().EmaCrossoverSetup == nil {
		return
	}

	params := strategies.Params{
		MarketData:     market.GetMarketData(),
		CandlesHandler: market.GetCandlesHandler(),
		Market:         market,
	}

	if market.GetMarketData().EmaCrossoverSetup.LongSetupParams != nil {
		market.Log("Calling EmaCrossoverLongs strategy")
		params.Type = ibroker.LongSide
		params.MarketStrategyParams = market.GetMarketData().EmaCrossoverSetup.LongSetupParams
		EmaCrossoverLongs(params)
	}

	if market.GetMarketData().EmaCrossoverSetup.ShortSetupParams != nil {
		market.Log("Calling EmaCrossoverShorts strategy")
		params.Type = ibroker.ShortSide
		params.MarketStrategyParams = market.GetMarketData().EmaCrossoverSetup.ShortSetupParams
		EmaCrossoverShorts(params)
	}
}
