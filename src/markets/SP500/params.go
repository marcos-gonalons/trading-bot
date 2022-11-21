package SP500

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 20 * priceAdjustment,
	MaxAttemptsToGetSL:  10,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 21,
		Past:   15,
	},
	MinStopLossDistance:              0 * priceAdjustment,
	MaxStopLossDistance:              600 * priceAdjustment,
	TakeProfitDistance:               160 * priceAdjustment,
	MinProfit:                        100 * priceAdjustment,
	CandlesAmountWithoutEMAsCrossing: 3,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 10 * priceAdjustment,
	MaxAttemptsToGetSL:  10,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 20,
		Past:   5,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 150 * priceAdjustment,
	TakeProfitDistance:  290 * priceAdjustment,
	MinProfit:           -40 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 240 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: -10 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 3,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}
