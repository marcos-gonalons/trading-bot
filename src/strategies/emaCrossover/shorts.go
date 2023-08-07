package emaCrossover

import (
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/candlesHandler/indicators/movingAverage"

	"TradingBot/src/markets"
	"TradingBot/src/strategies"
	"TradingBot/src/utils"
	"time"
)

func EmaCrossoverShorts(params strategies.Params) {
	var log = func(m string) {
		params.Market.Log("EmaCrossoverShorts | " + m)
	}
	container := params.Market.GetContainer()

	log("EmaCrossoverShorts started")
	defer func() {
		log("EmaCrossoverShorts ended")
	}()

	err := strategies.OnBegin(params)
	if err != nil {
		log(err.Error() + utils.GetStringRepresentation(params.MarketStrategyParams))
		return
	}

	candles := params.CandlesHandler.GetCompletedCandles()
	lastCompletedCandleIndex := len(candles) - 1
	lastCompletedCandle := candles[lastCompletedCandleIndex]

	openPosition := utils.FindPositionByMarket(container.APIData.GetPositions(), params.MarketData.BrokerAPIName)
	if openPosition != nil && container.API.IsShortPosition(openPosition) {
		closePositionOnReversal(
			openPosition,
			lastCompletedCandle,
			params.MarketStrategyParams.MinProfit,
			container.API,
			container.APIRetryFacade,
			params.MarketData,
			params.Market.Log,
			container.EmaService,
		)

		log("There is an open position - doing nothing ...")
		return
	}

	if container.EmaService.IsPriceAboveEMA(lastCompletedCandle, movingAverage.BASE_EMA) {
		log("Price is above huge EMA, not opening any shorts just yet ...")
		return
	}

	log("Price is below huge EMA, only shorts allowed...")

	for i := lastCompletedCandleIndex - params.MarketStrategyParams.EmaCrossover.CandlesAmountWithoutEMAsCrossing - 1; i < lastCompletedCandleIndex; i++ {
		if i <= 0 {
			return
		}

		if container.EmaService.GetEma(candles[i], movingAverage.SMALL_EMA).Value <= container.EmaService.GetEma(candles[i], movingAverage.BIG_EMA).Value {
			log("Small EMA was below the big EMA very recently - doing nothing - " + utils.GetStringRepresentation(lastCompletedCandle))
			return
		}
	}

	if container.EmaService.GetEma(lastCompletedCandle, movingAverage.SMALL_EMA).Value > container.EmaService.GetEma(lastCompletedCandle, movingAverage.BIG_EMA).Value {
		log("Small EMA is still above the big EMA - doing nothing - " + utils.GetStringRepresentation(lastCompletedCandle))
		return
	}

	price := lastCompletedCandle.Close

	stopLoss := getStopLoss(GetStopLossParams{
		LongOrShort:                     "short",
		PositionPrice:                   price,
		MinStopLossDistance:             params.MarketStrategyParams.MinStopLossDistance,
		MaxStopLossDistance:             params.MarketStrategyParams.MaxStopLossDistance,
		PriceOffset:                     params.MarketStrategyParams.EmaCrossover.StopLossPriceOffset,
		CandlesAmountForHorizontalLevel: params.MarketStrategyParams.CandlesAmountForHorizontalLevel,
		Candles:                         params.CandlesHandler.GetCompletedCandles(),
		GetHorizontalLevel:              container.HorizontalLevelsService.GetResistance,
		MaxAttempts:                     params.MarketStrategyParams.EmaCrossover.MaxAttemptsToGetSL,
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
		StopLossDistance:   stopLoss - price,
		TakeProfitDistance: params.MarketStrategyParams.TakeProfitDistance,
		RiskPercentage:     params.MarketStrategyParams.RiskPercentage,
		IsValidTime: utils.IsExecutionTimeValid(
			time.Now(),
			validMonths,
			validWeekdays,
			validHalfHours,
		),
		Side:                 ibroker.ShortSide,
		WithPendingOrders:    false,
		OrderType:            ibroker.MarketType,
		MinPositionSize:      params.MarketData.MinPositionSize,
		PositionSizeStrategy: params.MarketStrategyParams.PositionSizeStrategy,
	})
}
