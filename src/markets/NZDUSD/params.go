package NZDUSD

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1) / float32(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(0 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 20,
		Past:   0,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 900 * priceAdjustment,
	TakeProfitDistance:  200 * priceAdjustment,
	MinProfit:           99999 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: float64(15 * priceAdjustment),
		SLDistanceWhenTPIsVeryClose: float64(15 * priceAdjustment),
	},
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: float64(140 * priceAdjustment),
		TPDistanceWhenSLIsVeryClose: float64(-100 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 2,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(60 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 15,
		Past:   28,
	},
	MinStopLossDistance: 90 * priceAdjustment,
	MaxStopLossDistance: 900 * priceAdjustment,
	TakeProfitDistance:  500 * priceAdjustment,
	MinProfit:           370 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: float64(100 * priceAdjustment),
		SLDistanceWhenTPIsVeryClose: float64(40 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 27,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
