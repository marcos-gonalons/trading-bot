package NZDUSD

import "TradingBot/src/types"

var priceAdjustment float32 = 1 / 10000

var ResistanceBounceParams = types.StrategyParams{
	RiskPercentage:                  .5,
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
}

var SupportBounceParams = types.StrategyParams{
	RiskPercentage:                  .5,
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
}
