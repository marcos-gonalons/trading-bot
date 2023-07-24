package EURUSD

import (
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var priceAdjustment float64 = float64(1) / float64(10000)

var EMACrossoverLongParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 30,
		Past:   45,
	},
	MinStopLossDistance: 10 * priceAdjustment,
	MaxStopLossDistance: 600 * priceAdjustment,
	TakeProfitDistance:  230 * priceAdjustment,
	MinProfit:           99999 * priceAdjustment,
	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 30 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: -90 * priceAdjustment,
	},
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 100 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: -20 * priceAdjustment,
	},
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              75 * priceAdjustment,
		MaxAttemptsToGetSL:               12,
		CandlesAmountWithoutEMAsCrossing: 12,
	},
}

var EMACrossoverShortParams = types.MarketStrategyParams{
	RiskPercentage: 3,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 40,
		Past:   20,
	},
	MinStopLossDistance: 0 * priceAdjustment,
	MaxStopLossDistance: 600 * priceAdjustment,
	TakeProfitDistance:  350 * priceAdjustment,
	MinProfit:           120 * priceAdjustment,
	TrailingTakeProfit: &types.TrailingTakeProfit{
		SLDistanceShortForTighterTP: 40 * priceAdjustment,
		TPDistanceWhenSLIsVeryClose: -100 * priceAdjustment,
	},
	MaxSecondsOpenTrade: 0,

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,

	EmaCrossover: types.EmaCrossover{
		StopLossPriceOffset:              150 * priceAdjustment,
		MaxAttemptsToGetSL:               5,
		CandlesAmountWithoutEMAsCrossing: 0,
	},
}
