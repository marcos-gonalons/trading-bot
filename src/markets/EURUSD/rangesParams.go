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
	TakeProfitDistance:  120 * priceAdjustment,
	StopLossDistance:    25 * priceAdjustment,

	Ranges: types.Ranges{
		CandlesToCheck:                           300,
		MaxPriceDifferenceForSameHorizontalLevel: 25 * priceAdjustment,
		MinPriceDifferenceBetweenRangePoints:     80 * priceAdjustment,
		MinCandlesBetweenRangePoints:             5,
		MaxCandlesBetweenRangePoints:             300,
		PriceOffset:                              0,
		RangePoints:                              3,
		StartWith:                                types.RESISTANCE_TYPE,
		TakeProfitStrategy:                       "distance",
		StopLossStrategy:                         "level-with-offset",
		OrderType:                                constants.StopType,
		TrendyOnly:                               true,
	},

	MaxTradeExecutionPriceDifference: 9999,
	PositionSizeStrategy:             positionSize.BASED_ON_MULTIPLIER,
}

var RangesShortParams = types.MarketStrategyParams{}
