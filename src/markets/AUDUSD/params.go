package AUDUSD

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 2,
		Past:   0,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 200 * priceAdjustment,
	TakeProfitDistance:  200 * priceAdjustment,
	MinProfit:           10 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 120 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: 15 * priceAdjustment,
	},
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              0 * priceAdjustment,
		MaxAttemptsToGetSL:               8,
		CandlesAmountWithoutEMAsCrossing: 0,
	},
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 25,
		Past:   50,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 550 * priceAdjustment,
	TakeProfitDistance:  170 * priceAdjustment,
	MinProfit:           100 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 100 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: 60 * priceAdjustment,
	},
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              200 * priceAdjustment,
		MaxAttemptsToGetSL:               2,
		CandlesAmountWithoutEMAsCrossing: 6,
	},
}
