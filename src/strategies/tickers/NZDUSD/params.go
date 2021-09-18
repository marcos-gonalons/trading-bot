package NZDUSD

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1) / float32(10000)

/*
var ResistanceBounceParams = types.TickerStrategyParams{
	RiskPercentage:                  5,
	StopLossDistance:                0 * priceAdjustment,
	TakeProfitDistance:              0 * priceAdjustment,
	TPDistanceShortForTighterSL:     0 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     0 * float64(priceAdjustment),
	TrendCandles:                    0,
	TrendDiff:                       0 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 0,
	PriceOffset:                     0 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             0 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize: 10000,
}*/

var SupportBounceParams = types.TickerStrategyParams{
	RiskPercentage:                  5,
	StopLossDistance:                60 * priceAdjustment,
	TakeProfitDistance:              280 * priceAdjustment,
	TPDistanceShortForTighterSL:     110 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     -10 * float64(priceAdjustment),
	TrendCandles:                    0,
	TrendDiff:                       0 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 30,
	PriceOffset:                     -20 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             40 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
