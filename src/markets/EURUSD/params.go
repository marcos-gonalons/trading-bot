package EURUSD

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1) / float32(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(75 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 30,
		Past:   40,
	},
	MinStopLossDistance: 50 * priceAdjustment,
	MaxStopLossDistance: 580 * priceAdjustment,
	TakeProfitDistance:  230 * priceAdjustment,
	MinProfit:           99999 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: float64(30 * priceAdjustment),
		SLDistanceWhenTPIsVeryClose: float64(-90 * priceAdjustment),
	},
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: float64(100 * priceAdjustment),
		TPDistanceWhenSLIsVeryClose: float64(-20 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 12,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(150 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 50,
		Past:   15,
	},
	MinStopLossDistance: 20 * priceAdjustment,
	MaxStopLossDistance: 600 * priceAdjustment,
	TakeProfitDistance:  350 * priceAdjustment,
	MinProfit:           120 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: float64(40 * priceAdjustment),
		TPDistanceWhenSLIsVeryClose: float64(-100 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 0,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
