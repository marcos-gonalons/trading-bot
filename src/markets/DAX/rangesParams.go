package DAX

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
	MaxStopLossDistance: 300 * priceAdjustment,
	TakeProfitDistance:  80 * priceAdjustment,
	StopLossDistance:    -40 * priceAdjustment,

	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 0 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: 0 * priceAdjustment,
	},

	Ranges: types.Ranges{
		CandlesToCheck:                           400,
		MaxPriceDifferenceForSameHorizontalLevel: 150 * priceAdjustment,
		MinPriceDifferenceBetweenRangePoints:     800 * priceAdjustment,
		MinCandlesBetweenRangePoints:             3,
		MaxCandlesBetweenRangePoints:             300,
		MinimumDistanceToLevel:                   180 * priceAdjustment,
		PriceOffset:                              0 * priceAdjustment,
		RangePoints:                              3,
		StartWith:                                types.RESISTANCE_TYPE,
		TakeProfitStrategy:                       "level-with-offset",
		StopLossStrategy:                         "level-with-offset",
		OrderType:                                constants.MarketType,
		TrendyOnly:                               false,
	},

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}

var RangesShortParams = types.MarketStrategyParams{
	RiskPercentage: 1,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 35,
		Past:   35,
	},
	MaxStopLossDistance: 100 * priceAdjustment,
	TakeProfitDistance:  20 * priceAdjustment,
	StopLossDistance:    0 * priceAdjustment,

	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 0 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: 0 * priceAdjustment,
	},

	Ranges: types.Ranges{
		CandlesToCheck:                           400,
		MaxPriceDifferenceForSameHorizontalLevel: 400 * priceAdjustment,
		MinPriceDifferenceBetweenRangePoints:     300 * priceAdjustment,
		MinCandlesBetweenRangePoints:             35,
		MaxCandlesBetweenRangePoints:             500,
		MinimumDistanceToLevel:                   150 * priceAdjustment,
		PriceOffset:                              0 * priceAdjustment,
		RangePoints:                              3,
		StartWith:                                types.SUPPORT_TYPE,
		TakeProfitStrategy:                       "level-with-offset",
		StopLossStrategy:                         "level-with-offset",
		OrderType:                                constants.MarketType,
		TrendyOnly:                               true,
	},

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}
