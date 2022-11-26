package NZDUSD

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	StopLossPriceOffset: 0 * priceAdjustment,
	MaxAttemptsToGetSL:  20,
	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 50,
		Past:   60,
	},
	MinStopLossDistance:              40 * priceAdjustment,
	MaxStopLossDistance:              200 * priceAdjustment,
	TakeProfitDistance:               100 * priceAdjustment,
	MinProfit:                        99999 * priceAdjustment,
	CandlesAmountWithoutEMAsCrossing: 2,
	MaxSecondsOpenTrade:              0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}
