package USDCAD

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1) / float32(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(-280 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 35,
		Past:   20,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 500 * priceAdjustment,
	TakeProfitDistance:  575 * priceAdjustment,
	MinProfit:           250 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: float64(225 * priceAdjustment),
		SLDistanceWhenTPIsVeryClose: float64(75 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 21,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(25 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 25,
		Past:   45,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 1000 * priceAdjustment,
	TakeProfitDistance:  330 * priceAdjustment,
	MinProfit:           220 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: float64(40 * priceAdjustment),
		TPDistanceWhenSLIsVeryClose: float64(-180 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 21,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
