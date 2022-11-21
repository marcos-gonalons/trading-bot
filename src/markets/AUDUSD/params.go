package AUDUSD

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 0 * priceAdjustment,
	MaxAttemptsToGetSL:  10,
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
	CandlesAmountWithoutEMAsCrossing: 0,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 200 * priceAdjustment,
	MaxAttemptsToGetSL:  10,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 25,
		Past:   40,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 550 * priceAdjustment,
	TakeProfitDistance:  170 * priceAdjustment,
	MinProfit:           100 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 100 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: 60 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 6,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}
