package USDCHF

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1) / float32(10000)

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: float64(25 * priceAdjustment),
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 25,
		Past:   10,
	},
	MinStopLossDistance: 60 * priceAdjustment,
	MaxStopLossDistance: 500 * priceAdjustment,
	TakeProfitDistance:  120 * priceAdjustment,
	MinProfit:           60 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: float64(20 * priceAdjustment),
		TPDistanceWhenSLIsVeryClose: float64(-100 * priceAdjustment),
	},
	CandlesAmountWithoutEMAsCrossing: 10,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
