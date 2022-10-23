package emaCrossover

import (
	ibroker "TradingBot/src/services/api/ibroker/constants"

	"TradingBot/src/markets"
	"TradingBot/src/strategies"
	"TradingBot/src/utils"
	"time"
)

func EmaCrossoverLongs(params strategies.Params) {
	params.Market.Log("EmaCrossoverLongs started")
	defer func() {
		params.Market.Log("EmaCrossoverLongs ended")
	}()

	err := strategies.OnBegin(params)
	if err != nil {
		params.Market.Log(err.Error() + utils.GetStringRepresentation(params.MarketStrategyParams))
		return
	}

	candles := params.CandlesHandler.GetCandles()
	lastCompletedCandleIndex := len(candles) - 2
	lastCompletedCandle := candles[lastCompletedCandleIndex]

	openPosition := utils.FindPositionByMarket(params.Container.APIData.GetPositions(), params.MarketData.BrokerAPIName)
	if openPosition != nil {
		closePositionOnReversal(
			openPosition,
			lastCompletedCandle,
			params.MarketStrategyParams.MinProfit,
			params.Container.API,
		)

		params.Market.Log("There is an open position - doing nothing ...")
		return
	}

	if lastCompletedCandle.Close <= getEma(lastCompletedCandle, BASE_EMA).Value {
		params.Market.Log("Price is below huge EMA, not opening any longs just yet ...")
		return
	}

	params.Market.Log("Price is above huge EMA, only longs allowed...")

	for i := lastCompletedCandleIndex - params.MarketStrategyParams.CandlesAmountWithoutEMAsCrossing - 1; i < lastCompletedCandleIndex; i++ {
		if i <= 0 {
			return
		}

		if getEma(candles[i], SMALL_EMA).Value >= getEma(candles[i], BIG_EMA).Value {
			params.Market.Log("Small EMA was above the big EMA very recently - doing nothing - " + utils.GetStringRepresentation(lastCompletedCandle))
			return
		}
	}

	if getEma(candles[lastCompletedCandleIndex], SMALL_EMA).Value < getEma(candles[lastCompletedCandleIndex], BIG_EMA).Value {
		params.Market.Log("Small EMA is still below the big EMA - doing nothing - " + utils.GetStringRepresentation(lastCompletedCandle))
		return
	}

	price := params.Market.GetCurrentBrokerQuote().Ask

	stopLoss := getStopLoss(GetStopLossParams{
		LongOrShort:                     "long",
		PositionPrice:                   price,
		MinStopLossDistance:             params.MarketStrategyParams.MinStopLossDistance,
		MaxStopLossDistance:             params.MarketStrategyParams.MaxStopLossDistance,
		CandleIndex:                     lastCompletedCandleIndex,
		PriceOffset:                     float32(params.MarketStrategyParams.StopLossPriceOffset),
		CandlesAmountForHorizontalLevel: params.MarketData.LongSetupParams.CandlesAmountForHorizontalLevel,
		Candles:                         params.CandlesHandler.GetCandles(),
		GetResistancePrice:              params.Container.HorizontalLevelsService.GetResistancePrice,
		GetSupportPrice:                 params.Container.HorizontalLevelsService.GetSupportPrice,
	})

	params.Market.OnValidTradeSetup(markets.OnValidTradeSetupParams{
		Price:              float64(price),
		StopLossDistance:   float32(float64(price) - stopLoss),
		TakeProfitDistance: params.MarketStrategyParams.TakeProfitDistance,
		RiskPercentage:     params.MarketStrategyParams.RiskPercentage,
		IsValidTime: utils.IsExecutionTimeValid(
			time.Now(),
			params.MarketStrategyParams.ValidTradingTimes.ValidMonths,
			params.MarketStrategyParams.ValidTradingTimes.ValidWeekdays,
			params.MarketStrategyParams.ValidTradingTimes.ValidHalfHours,
		),
		Side:              ibroker.LongSide,
		WithPendingOrders: false,
		OrderType:         ibroker.MarketType,
		MinPositionSize:   params.MarketStrategyParams.MinPositionSize,
	})
}
