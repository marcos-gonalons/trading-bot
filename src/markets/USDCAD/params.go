package USDCAD

import "TradingBot/src/types"

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: -280 * priceAdjustment,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 35,
		Past:   20,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 500 * priceAdjustment,
	TakeProfitDistance:  575 * priceAdjustment,
	MinProfit:           250 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 225 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: 75 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 21,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 25 * priceAdjustment,
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
	CandlesAmountWithoutEMAsCrossing: 21,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
}
