package DAX

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(-75 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 23,
		Past:   30,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 1300 * priceAdjustment,
	TakeProfitDistance:  450 * priceAdjustment,
	MinProfit:           250 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: float64(120 * priceAdjustment),
		SLDistanceWhenTPIsVeryClose: float64(-75 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 3,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  1,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(-125 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 30,
		Past:   20,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 1000 * priceAdjustment,
	TakeProfitDistance:  450 * priceAdjustment,
	MinProfit:           50 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: float64(220 * priceAdjustment),
		SLDistanceWhenTPIsVeryClose: float64(-60 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 6,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  1,
}
