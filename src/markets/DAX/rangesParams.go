package DAX

import (
	"TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
)

var RangesLongParams = types.MarketStrategyParams{
	RiskPercentage: 1,

	CandlesAmountForHorizontalLevel: &types.CandlesAmountForHorizontalLevel{
		Future: 10,
		Past:   20,
	},
	MaxStopLossDistance: 300 * priceAdjustment,
	TakeProfitDistance:  60 * priceAdjustment,
	StopLossDistance:    -10 * priceAdjustment,

	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 70 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: -40 * priceAdjustment,
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
		Future: 3,
		Past:   25,
	},
	MaxStopLossDistance: 300 * priceAdjustment,
	TakeProfitDistance:  120 * priceAdjustment,
	StopLossDistance:    -10 * priceAdjustment,

	TrailingStopLoss: &types.TrailingStopLoss{
		TPDistanceShortForTighterSL: 0 * priceAdjustment,
		SLDistanceWhenTPIsVeryClose: 0 * priceAdjustment,
	},

	Ranges: types.Ranges{
		CandlesToCheck:                           400,
		MaxPriceDifferenceForSameHorizontalLevel: 25 * priceAdjustment,
		MinPriceDifferenceBetweenRangePoints:     150 * priceAdjustment,
		MinCandlesBetweenRangePoints:             5,
		MaxCandlesBetweenRangePoints:             300,
		MinimumDistanceToLevel:                   40 * priceAdjustment,
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
