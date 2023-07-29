package ranges

import (
	"TradingBot/src/services"
	"TradingBot/src/services/candlesHandler/indicators/movingAverage"
	"TradingBot/src/strategies"
	"TradingBot/src/utils"
)

func RangesShorts(params strategies.Params) {
	var log = func(m string) {
		params.Market.Log("RangesShorts | " + m)
	}

	log("RangesShorts started")
	defer func() {
		log("RangesShorts ended")
	}()

	err := strategies.OnBegin(params)
	if err != nil {
		log(err.Error() + utils.GetStringRepresentation(params.MarketStrategyParams))
		return
	}
	candles := params.CandlesHandler.GetCompletedCandles()
	lastCompletedCandleIndex := len(candles) - 1
	lastCompletedCandle := candles[lastCompletedCandleIndex]

	container := services.GetServicesContainer()

	openPosition := utils.FindPositionByMarket(container.APIData.GetPositions(), params.MarketData.BrokerAPIName)
	if openPosition != nil && container.API.IsShortPosition(openPosition) {
		log("There is an open position - doing nothing ...")
		return
	}

	if params.MarketStrategyParams.Ranges.TrendyOnly {
		if container.EmaService.IsPriceAboveEMA(lastCompletedCandle, movingAverage.BASE_EMA) {
			log("TrendyOnly flag is active and the current price is above the base EMA - Doing nothing ...")
			return
		}
	}

}
