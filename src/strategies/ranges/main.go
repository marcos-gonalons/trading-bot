package ranges

import (
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/strategies"
)

func OnNewCandle(market markets.MarketInterface) {
	if market.GetMarketData().RangesSetup == nil {
		return
	}

	params := strategies.Params{
		MarketData:     market.GetMarketData(),
		CandlesHandler: market.GetCandlesHandler(),
		Market:         market,
	}

	if market.GetMarketData().RangesSetup.LongSetupParams != nil {
		market.Log("Calling RangesLongs strategy")
		params.Type = ibroker.LongSide
		params.MarketStrategyParams = market.GetMarketData().RangesSetup.LongSetupParams
		RangesLongs(params)
	}

	if market.GetMarketData().RangesSetup.ShortSetupParams != nil {
		market.Log("Calling RangesShorts strategy")
		params.Type = ibroker.ShortSide
		params.MarketStrategyParams = market.GetMarketData().RangesSetup.ShortSetupParams
		RangesShorts(params)
	}
}
