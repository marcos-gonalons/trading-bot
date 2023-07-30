package EURUSD

import (
	"TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var RangesLongParams = types.MarketStrategyParams{
	RiskPercentage: 1,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 10,
		Past:   10,
	},
	MaxStopLossDistance: 340 * priceAdjustment,
	TakeProfitDistance:  100 * priceAdjustment,
	StopLossDistance:    50 * priceAdjustment,

	Ranges: types.Ranges{
		CandlesToCheck:                           300,
		MaxPriceDifferenceForSameHorizontalLevel: 70 * priceAdjustment,
		MinPriceDifferenceBetweenRangePoints:     100 * priceAdjustment,
		MinCandlesBetweenRangePoints:             10,
		MaxCandlesBetweenRangePoints:             100,
		PriceOffset:                              0,
		RangePoints:                              3,
		StartWith:                                types.RESISTANCE_TYPE,
		TakeProfitStrategy:                       "distance",
		StopLossStrategy:                         "level-with-offset",
		OrderType:                                constants.StopType,
		TrendyOnly:                               false,
	},

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}

var RangesShortParams = types.MarketStrategyParams{}
