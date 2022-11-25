package USDCHF

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 25 * priceAdjustment,
	MaxAttemptsToGetSL:  11,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 30,
		Past:   45,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 200 * priceAdjustment,
	TakeProfitDistance:  120 * priceAdjustment,
	MinProfit:           60 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 20 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: -100 * priceAdjustment,
	},
	CandlesAmountWithoutEMAsCrossing: 10,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}
