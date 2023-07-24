package USDCAD

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 35,
		Past:   25,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 500 * priceAdjustment,
	TakeProfitDistance:  550 * priceAdjustment,
	MinProfit:           250 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 225 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: 75 * priceAdjustment,
	},
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              -280 * priceAdjustment,
		MaxAttemptsToGetSL:               2,
		CandlesAmountWithoutEMAsCrossing: 21,
	},
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 25,
		Past:   45,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 680 * priceAdjustment,
	TakeProfitDistance:  330 * priceAdjustment,
	MinProfit:           220 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 40 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: -180 * priceAdjustment,
	},
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              25 * priceAdjustment,
		MaxAttemptsToGetSL:               10,
		CandlesAmountWithoutEMAsCrossing: 21,
	},
}
