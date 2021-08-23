package USDCAD

import "TradingBot/src/types"

var priceAdjustment float32 = 1 / 10000

var ResistanceBounceParams = types.StrategyParams{
	RiskPercentage:                  1.5,
	StopLossDistance:                200 * priceAdjustment,
	TakeProfitDistance:              260 * priceAdjustment,
	TPDistanceShortForTighterSL:     230 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     170 * float64(priceAdjustment),
	TrendCandles:                    132,
	TrendDiff:                       50 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 30,
	PriceOffset:                     30 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             42 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
}

var SupportBounceParams = types.StrategyParams{
	RiskPercentage:                  1.5,
	StopLossDistance:                340 * priceAdjustment,
	TakeProfitDistance:              60 * priceAdjustment,
	TPDistanceShortForTighterSL:     10 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     -190 * float64(priceAdjustment),
	TrendCandles:                    12,
	TrendDiff:                       40 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 30,
	PriceOffset:                     36 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             50 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
}
