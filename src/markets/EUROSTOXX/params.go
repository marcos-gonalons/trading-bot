package EUROSTOXX

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 35,
		Past:   10,
	},
	MinStopLossDistance: 100 * priceAdjustment,
	MaxStopLossDistance: 600 * priceAdjustment,
	TakeProfitDistance:  60 * priceAdjustment,
	MinProfit:           40 * priceAdjustment,
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              90 * priceAdjustment,
		MaxAttemptsToGetSL:               2,
		CandlesAmountWithoutEMAsCrossing: 0,
	},
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 20,
		Past:   3,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 500 * priceAdjustment,
	TakeProfitDistance:  45 * priceAdjustment,
	MinProfit:           99999 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 20 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: -30 * priceAdjustment,
	},
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              125 * priceAdjustment,
		MaxAttemptsToGetSL:               2,
		CandlesAmountWithoutEMAsCrossing: 3,
	},
}
