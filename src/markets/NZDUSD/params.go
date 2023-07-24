package NZDUSD

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 50,
		Past:   60,
	},
	MinStopLossDistance: 40 * priceAdjustment,
	MaxStopLossDistance: 200 * priceAdjustment,
	TakeProfitDistance:  100 * priceAdjustment,
	MinProfit:           99999 * priceAdjustment,
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              0 * priceAdjustment,
		MaxAttemptsToGetSL:               20,
		CandlesAmountWithoutEMAsCrossing: 2,
	},
}
