package EUROSTOXX

import "TradingBot/src/types"

var priceAdjustment float64 = float64(1)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 90 * priceAdjustment,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 12,
		Past:   0,
	},
	MinStopLossDistance:              0 * priceAdjustment,
	MaxStopLossDistance:              950 * priceAdjustment,
	TakeProfitDistance:               60 * priceAdjustment,
	MinProfit:                        9999 * priceAdjustment,
	CandlesAmountWithoutEMAsCrossing: 0,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  1,
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 125 * priceAdjustment,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 30,
		Past:   0,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 950 * priceAdjustment,
	TakeProfitDistance:  40 * priceAdjustment,
	MinProfit:           99999 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 20 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: -30 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 3,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  1,
}
