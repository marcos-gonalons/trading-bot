package NZDUSD

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 0 * priceAdjustment,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 20,
		Past:   0,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 300 * priceAdjustment,
	TakeProfitDistance:  200 * priceAdjustment,
	MinProfit:           99999 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 15 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: 15 * priceAdjustment,
	},
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 140 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: -100 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 2,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 60 * priceAdjustment,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 15,
		Past:   28,
	},
	MinStopLossDistance: 90 * priceAdjustment,
	MaxStopLossDistance: 620 * priceAdjustment,
	TakeProfitDistance:  500 * priceAdjustment,
	MinProfit:           370 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 100 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: 40 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 27,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}
