package DAX

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 1,

	StopLossPriceOffset: float64(20 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 0,
		Past:   0,
	},
	MinStopLossDistance:              0 * priceAdjustment,
	MaxStopLossDistance:              900 * priceAdjustment,
	TakeProfitDistance:               160 * priceAdjustment,
	MinProfit:                        100 * priceAdjustment,
	CandlesAmountWithoutEMAsCrossing: 3,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  1,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 1,

	StopLossPriceOffset: float64(10 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 20,
		Past:   5,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 150 * priceAdjustment,
	TakeProfitDistance:  290 * priceAdjustment,
	MinProfit:           -40 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: float64(240 * priceAdjustment),
		SLDistanceWhenTPIsVeryClose: float64(-10 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 3,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  1,
}
