package USDCHF

import "TradingBot/src/types"

var priceAdjustment float32 = 1 / 10000

var ResistanceBreakoutParams = types.StrategyParams{
	RiskPercentage:                  .5,
	StopLossDistance:                80 * priceAdjustment,
	TakeProfitDistance:              130 * priceAdjustment,
	TPDistanceShortForTighterSL:     10 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     60 * float64(priceAdjustment),
	TrendCandles:                    110,
	TrendDiff:                       35 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 8,
	PriceOffset:                     60 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             12 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
}

var SupportBreakoutParams = types.StrategyParams{
	RiskPercentage:                  .5,
	StopLossDistance:                200 * priceAdjustment,
	TakeProfitDistance:              120 * priceAdjustment,
	TPDistanceShortForTighterSL:     0 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     0 * float64(priceAdjustment),
	TrendCandles:                    180,
	TrendDiff:                       70 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 50,
	PriceOffset:                     30 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             18 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
}
