package emaCrossover

import (
	ibroker "TradingBot/src/services/api/ibroker/constants"

	"TradingBot/src/markets"
	"TradingBot/src/strategies"
	"TradingBot/src/utils"
	"time"
)

func EmaCrossoverShorts(params strategies.Params) {
	var log = func(m string) {
		params.Market.Log("EmaCrossoverShorts | " + m)
	}

	log("EmaCrossoverShorts started")
	defer func() {
		log("EmaCrossoverShorts ended")
	}()

	err := strategies.OnBegin(params)
	if err != nil {
		log(err.Error() + utils.GetStringRepresentation(params.MarketStrategyParams))
		return
	}

	candles := params.CandlesHandler.GetCandles()
	lastCompletedCandleIndex := len(candles) - 2
	lastCompletedCandle := candles[lastCompletedCandleIndex]

	openPosition := utils.FindPositionByMarket(params.Container.APIData.GetPositions(), params.MarketData.BrokerAPIName)
	if openPosition != nil && params.Container.API.IsShortPosition(openPosition) {
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

	if lastCompletedCandle.Close >= getEma(lastCompletedCandle, BASE_EMA).Value {
		log("Price is above huge EMA, not opening any shorts just yet ...")
		return
	}

	log("Price is below huge EMA, only shorts allowed...")

	for i := lastCompletedCandleIndex - params.MarketStrategyParams.CandlesAmountWithoutEMAsCrossing - 1; i < lastCompletedCandleIndex; i++ {
		if i <= 0 {
			return
		}

		if getEma(candles[i], SMALL_EMA).Value <= getEma(candles[i], BIG_EMA).Value {
			log("Small EMA was below the big EMA very recently - doing nothing - " + utils.GetStringRepresentation(lastCompletedCandle))
			return
		}
	}

	if getEma(candles[lastCompletedCandleIndex], SMALL_EMA).Value > getEma(candles[lastCompletedCandleIndex], BIG_EMA).Value {
		log("Small EMA is still above the big EMA - doing nothing - " + utils.GetStringRepresentation(lastCompletedCandle))
		return
	}

	price := params.Market.GetCurrentBrokerQuote().Ask

	stopLoss := getStopLoss(GetStopLossParams{
		LongOrShort:                     "short",
		PositionPrice:                   price,
		MinStopLossDistance:             params.MarketStrategyParams.MinStopLossDistance,
		MaxStopLossDistance:             params.MarketStrategyParams.MaxStopLossDistance,
		CandleIndex:                     lastCompletedCandleIndex,
		PriceOffset:                     params.MarketStrategyParams.StopLossPriceOffset,
		CandlesAmountForHorizontalLevel: params.MarketData.ShortSetupParams.CandlesAmountForHorizontalLevel,
		Candles:                         params.CandlesHandler.GetCandles(),
		GetResistancePrice:              params.Container.HorizontalLevelsService.GetResistancePrice,
		GetSupportPrice:                 params.Container.HorizontalLevelsService.GetSupportPrice,
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
