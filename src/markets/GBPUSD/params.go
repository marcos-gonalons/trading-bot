package GBPUSD

import "TradingBot/src/types"

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: -100 * priceAdjustment,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 30,
		Past:   30,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 800 * priceAdjustment,
	TakeProfitDistance:  450 * priceAdjustment,
	MinProfit:           300 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 105 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: -150 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 21,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 75 * priceAdjustment,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 3,
		Past:   6,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 600 * priceAdjustment,
	TakeProfitDistance:  130 * priceAdjustment,
	MinProfit:           40 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 60 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: -50 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 0,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
