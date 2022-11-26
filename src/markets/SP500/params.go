package SP500

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 20 * priceAdjustment,
	MaxAttemptsToGetSL:  2,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 21,
		Past:   15,
	},
	MinStopLossDistance:              120 * priceAdjustment,
	MaxStopLossDistance:              275 * priceAdjustment,
	TakeProfitDistance:               120 * priceAdjustment,
	MinProfit:                        60 * priceAdjustment,
	CandlesAmountWithoutEMAsCrossing: 3,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 10 * priceAdjustment,
	MaxAttemptsToGetSL:  2,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 20,
		Past:   5,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 150 * priceAdjustment,
	TakeProfitDistance:  280 * priceAdjustment,
	MinProfit:           -40 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 0 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: 0 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 3,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}
