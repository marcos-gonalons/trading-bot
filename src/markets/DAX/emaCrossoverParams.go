package DAX

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 23,
		Past:   30,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 500 * priceAdjustment,
	TakeProfitDistance:  450 * priceAdjustment,
	MinProfit:           250 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 120 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: -75 * priceAdjustment,
	},
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              -75 * priceAdjustment,
		MaxAttemptsToGetSL:               2,
		CandlesAmountWithoutEMAsCrossing: 3,
	},
}
