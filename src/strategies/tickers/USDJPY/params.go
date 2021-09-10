package USDJPY

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1) / float32(100)

var ResistanceBounceParams = types.StrategyParams{
	RiskPercentage:                  .5,
	StopLossDistance:                45 * priceAdjustment,
	TakeProfitDistance:              330 * priceAdjustment,
	TPDistanceShortForTighterSL:     0 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     0 * float64(priceAdjustment),
	TrendCandles:                    0,
	TrendDiff:                       0 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 10,
	PriceOffset:                     50 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             45 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
}

var SupportBounceParams = types.StrategyParams{
	RiskPercentage:                  .5,
	StopLossDistance:                190 * priceAdjustment,
	TakeProfitDistance:              70 * priceAdjustment,
	TPDistanceShortForTighterSL:     30 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     -110 * float64(priceAdjustment),
	TrendCandles:                    50,
	TrendDiff:                       10 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 45,
	PriceOffset:                     -40 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             25 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
}
