package GBPUSD

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 30,
		Past:   30,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 550 * priceAdjustment,
	TakeProfitDistance:  450 * priceAdjustment,
	MinProfit:           300 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 105 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: -150 * priceAdjustment,
	},

	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              -100 * priceAdjustment,
		MaxAttemptsToGetSL:               2,
		CandlesAmountWithoutEMAsCrossing: 21,
	},
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 3,
		Past:   6,
	},
	MinStopLossDistance: 40 * priceAdjustment,
	MaxStopLossDistance: 500 * priceAdjustment,
	TakeProfitDistance:  130 * priceAdjustment,
	MinProfit:           40 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 60 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: -50 * priceAdjustment,
	},
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              75 * priceAdjustment,
		MaxAttemptsToGetSL:               5,
		CandlesAmountWithoutEMAsCrossing: 0,
	},
}
