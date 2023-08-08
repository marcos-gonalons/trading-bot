package EURUSD

import (
	"TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var RangesLongParams = types.MarketStrategyParams{
	RiskPercentage: 1,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 5,
		Past:   5,
	},
	MaxStopLossDistance: 300 * priceAdjustment,
	TakeProfitDistance:  40 * priceAdjustment,
	StopLossDistance:    70 * priceAdjustment,

	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 0 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: 0 * priceAdjustment,
	},

	Ranges: types.Ranges{
		CandlesToCheck:                           400,
		MaxPriceDifferenceForSameHorizontalLevel: 25 * priceAdjustment,
		MinPriceDifferenceBetweenRangePoints:     120 * priceAdjustment,
		MinCandlesBetweenRangePoints:             5,
		MaxCandlesBetweenRangePoints:             300,
		MinimumDistanceToLevel:                   30 * priceAdjustment,
		PriceOffset:                              0 * priceAdjustment,
		RangePoints:                              3,
		StartWith:                                types.RESISTANCE_TYPE,
		TakeProfitStrategy:                       "level-with-offset",
		StopLossStrategy:                         "level-with-offset",
		OrderType:                                constants.MarketType,
		TrendyOnly:                               true,
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
		MinimumDistanceToLevel:                   80 * priceAdjustment,
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
