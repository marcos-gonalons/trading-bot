package GBPUSD

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1) / float32(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(-100 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 30,
		Past:   30,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 800 * priceAdjustment,
	TakeProfitDistance:  450 * priceAdjustment,
	MinProfit:           300 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: float64(105 * priceAdjustment),
		SLDistanceWhenTPIsVeryClose: float64(-150 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 21,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(75 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 3,
		Past:   6,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 600 * priceAdjustment,
	TakeProfitDistance:  130 * priceAdjustment,
	MinProfit:           40 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: float64(60 * priceAdjustment),
		TPDistanceWhenSLIsVeryClose: float64(-50 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 0,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
