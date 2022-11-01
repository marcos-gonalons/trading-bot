package USDCHF

import "TradingBot/src/types"

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 25 * priceAdjustment,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 25,
		Past:   10,
	},
	MinStopLossDistance: 60 * priceAdjustment,
	MaxStopLossDistance: 500 * priceAdjustment,
	TakeProfitDistance:  120 * priceAdjustment,
	MinProfit:           60 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 20 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: -100 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 10,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
