package ranges

import (
	"TradingBot/src/strategies"
	"TradingBot/src/utils"
)

func RangesLongs(params strategies.Params) {
	var log = func(m string) {
		params.Market.Log("RangesLongs | " + m)
	}

	log("RangesLongs started")
	defer func() {
		log("RangesLongs ended")
	}()

	err := strategies.OnBegin(params)
	if err != nil {
		log(err.Error() + utils.GetStringRepresentation(params.MarketStrategyParams))
		return
	}
	/*
		candles := params.CandlesHandler.GetCompletedCandles()
		lastCompletedCandleIndex := len(candles) - 1
		lastCompletedCandle := candles[lastCompletedCandleIndex]


		container := services.GetServicesContainer()

		openPosition := utils.FindPositionByMarket(container.APIData.GetPositions(), params.MarketData.BrokerAPIName)
		if openPosition != nil && container.API.IsLongPosition(openPosition) {
			log("There is an open position - doing nothing ...")
			return
		}

		if params.MarketStrategyParams.Ranges.TrendyOnly {
			services.GetServicesContainer().TrendsService()
		}
	*/

}
