package AUDUSD

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1) / float32(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(0 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 2,
		Past:   0,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 200 * priceAdjustment,
	TakeProfitDistance:  200 * priceAdjustment,
	MinProfit:           10 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: float64(120 * priceAdjustment),
		SLDistanceWhenTPIsVeryClose: float64(15 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 0,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(200 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 25,
		Past:   40,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 550 * priceAdjustment,
	TakeProfitDistance:  170 * priceAdjustment,
	MinProfit:           100 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: float64(100 * priceAdjustment),
		TPDistanceWhenSLIsVeryClose: float64(60 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 6,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
