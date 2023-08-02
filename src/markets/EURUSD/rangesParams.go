package EURUSD

import (
	"TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var RangesLongParams = types.MarketStrategyParams{
	RiskPercentage: 1,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 3,
		Past:   3,
	},
	MaxStopLossDistance: 450 * priceAdjustment,
	TakeProfitDistance:  50 * priceAdjustment,
	StopLossDistance:    -30 * priceAdjustment,

	Ranges: types.Ranges{
		CandlesToCheck:                           400,
		MaxPriceDifferenceForSameHorizontalLevel: 75 * priceAdjustment,
		MinPriceDifferenceBetweenRangePoints:     25 * priceAdjustment,
		MinCandlesBetweenRangePoints:             5,
		MaxCandlesBetweenRangePoints:             150,
		PriceOffset:                              -50 * priceAdjustment,
		RangePoints:                              3,
		StartWith:                                types.SUPPORT_TYPE,
		TakeProfitStrategy:                       "distance",
		StopLossStrategy:                         "level-with-offset",
		OrderType:                                constants.StopType,
		TrendyOnly:                               false,
	},

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}

var RangesShortParams = types.MarketStrategyParams{
	RiskPercentage: 1,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 2,
		Past:   2,
	},
	MaxStopLossDistance: 350 * priceAdjustment,
	TakeProfitDistance:  200 * priceAdjustment,
	StopLossDistance:    50 * priceAdjustment,

	Ranges: types.Ranges{
		CandlesToCheck:                           1000,
		MaxPriceDifferenceForSameHorizontalLevel: 25 * priceAdjustment,
		MinPriceDifferenceBetweenRangePoints:     20 * priceAdjustment,
		MinCandlesBetweenRangePoints:             4,
		MaxCandlesBetweenRangePoints:             100,
		PriceOffset:                              0 * priceAdjustment,
		RangePoints:                              3,
		StartWith:                                types.SUPPORT_TYPE,
		TakeProfitStrategy:                       "distance",
		StopLossStrategy:                         "half",
		OrderType:                                constants.StopType,
		TrendyOnly:                               false,
	},

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}
