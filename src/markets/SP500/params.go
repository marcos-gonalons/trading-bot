package SP500

import "TradingBot/src/types"

var priceAdjustment float64 = float64(1)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 20 * priceAdjustment,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 0,
		Past:   0,
	},
	MinStopLossDistance:              0 * priceAdjustment,
	MaxStopLossDistance:              600 * priceAdjustment,
	TakeProfitDistance:               160 * priceAdjustment,
	MinProfit:                        100 * priceAdjustment,
	CandlesAmountWithoutEMAsCrossing: 3,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 10 * priceAdjustment,
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
}
