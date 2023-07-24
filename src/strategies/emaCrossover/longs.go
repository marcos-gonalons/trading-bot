package emaCrossover

import (
	ibroker "TradingBot/src/services/api/ibroker/constants"

	"TradingBot/src/markets"
	"TradingBot/src/strategies"
	"TradingBot/src/utils"
	"time"
)

func EmaCrossoverLongs(params strategies.Params) {
	var log = func(m string) {
		params.Market.Log("EmaCrossoverLongs | " + m)
	}

	log("EmaCrossoverLongs started")
	defer func() {
		log("EmaCrossoverLongs ended")
	}()

	err := strategies.OnBegin(params)
	if err != nil {
		log(err.Error() + utils.GetStringRepresentation(params.MarketStrategyParams))
		return
	}

	candles := params.CandlesHandler.GetCompletedCandles()
	lastCompletedCandleIndex := len(candles) - 1
	lastCompletedCandle := candles[lastCompletedCandleIndex]

	openPosition := utils.FindPositionByMarket(params.Container.APIData.GetPositions(), params.MarketData.BrokerAPIName)
	if openPosition != nil && params.Container.API.IsLongPosition(openPosition) {
		closePositionOnReversal(
			openPosition,
			lastCompletedCandle,
			params.MarketStrategyParams.MinProfit,
			params.Container.API,
			params.Container.APIRetryFacade,
			params.MarketData,
			params.Market.Log,
		)

		log("There is an open position - doing nothing ...")
		return
	}

	if lastCompletedCandle.Close <= getEma(lastCompletedCandle, BASE_EMA).Value {
		log("Price is below huge EMA, not opening any longs just yet ...")
		return
	}

	log("Price is above huge EMA, only longs allowed...")

	for i := lastCompletedCandleIndex - params.MarketStrategyParams.CandlesAmountWithoutEMAsCrossing - 1; i < lastCompletedCandleIndex; i++ {
		if i <= 0 {
			return
		}

		if getEma(candles[i], SMALL_EMA).Value >= getEma(candles[i], BIG_EMA).Value {
			log("Small EMA was above the big EMA very recently - doing nothing - " + utils.GetStringRepresentation(lastCompletedCandle))
			return
		}
	}

	if getEma(lastCompletedCandle, SMALL_EMA).Value < getEma(lastCompletedCandle, BIG_EMA).Value {
		log("Small EMA is still below the big EMA - doing nothing - " + utils.GetStringRepresentation(lastCompletedCandle))
		return
	}

	price := lastCompletedCandle.Close

	stopLoss := getStopLoss(GetStopLossParams{
		LongOrShort:                     "long",
		PositionPrice:                   price,
		MinStopLossDistance:             params.MarketStrategyParams.MinStopLossDistance,
		MaxStopLossDistance:             params.MarketStrategyParams.MaxStopLossDistance,
		PriceOffset:                     params.MarketStrategyParams.StopLossPriceOffset,
		CandlesAmountForHorizontalLevel: params.MarketData.LongSetupParams.CandlesAmountForHorizontalLevel,
		Candles:                         params.CandlesHandler.GetCompletedCandles(),
		GetHorizontalLevel:              params.Container.HorizontalLevelsService.GetSupport,
		MaxAttempts:                     params.MarketStrategyParams.MaxAttemptsToGetSL,
		Log:                             params.Market.Log,
	})

	var validMonths, validWeekdays, validHalfHours []string
	if params.MarketStrategyParams.ValidTradingTimes != nil {
		validMonths = params.MarketStrategyParams.ValidTradingTimes.ValidMonths
		validWeekdays = params.MarketStrategyParams.ValidTradingTimes.ValidWeekdays
		validHalfHours = params.MarketStrategyParams.ValidTradingTimes.ValidHalfHours
	}

	params.Market.OnValidTradeSetup(markets.OnValidTradeSetupParams{
		Price:              price,
		StopLossDistance:   price - stopLoss,
		TakeProfitDistance: params.MarketStrategyParams.TakeProfitDistance,
		RiskPercentage:     params.MarketStrategyParams.RiskPercentage,
		IsValidTime: utils.IsExecutionTimeValid(
			time.Now(),
			validMonths,
			validWeekdays,
			validHalfHours,
		),
		Side:                 ibroker.LongSide,
		WithPendingOrders:    false,
		OrderType:            ibroker.MarketType,
		MinPositionSize:      params.MarketData.MinPositionSize,
		PositionSizeStrategy: params.MarketStrategyParams.PositionSizeStrategy,
	})
}
