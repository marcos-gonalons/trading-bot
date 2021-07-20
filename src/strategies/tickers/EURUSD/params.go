package EURUSD

import "TradingBot/src/types"

var priceAdjustment float32 = 1 / 10000

var ResistanceBounceParams = types.StrategyParams{
	RiskPercentage:                  1,
	StopLossDistance:                290.0 * priceAdjustment,
	TakeProfitDistance:              460 * priceAdjustment,
	CandlesAmountForHorizontalLevel: 27,
	TPDistanceShortForTighterSL:     100 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     -150 * float64(priceAdjustment),
	PriceOffset:                     -5 * float64(priceAdjustment),
	TrendCandles:                    120,
	TrendDiff:                       200 * float64(priceAdjustment),
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
}

var SupportBounceParams = types.StrategyParams{
	RiskPercentage:                  1,
	StopLossDistance:                180 * priceAdjustment,
	TakeProfitDistance:              370 * priceAdjustment,
	CandlesAmountForHorizontalLevel: 27,
	TPDistanceShortForTighterSL:     150 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     100 * float64(priceAdjustment),
	PriceOffset:                     -13 * float64(priceAdjustment),
	TrendCandles:                    72,
	TrendDiff:                       10 * float64(priceAdjustment),
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
}
