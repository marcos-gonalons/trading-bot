package EURUSD

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1) / float32(10000)

var ResistanceBounceParams = types.TickerStrategyParams{
	RiskPercentage:                  5,
	StopLossDistance:                290 * priceAdjustment,
	TakeProfitDistance:              460 * priceAdjustment,
	TPDistanceShortForTighterSL:     150 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     0 * float64(priceAdjustment),
	TrendCandles:                    120,
	TrendDiff:                       200 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 27,
	PriceOffset:                     -10 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             35 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}

var SupportBounceParams = types.TickerStrategyParams{
	RiskPercentage:                  5,
	StopLossDistance:                180 * priceAdjustment,
	TakeProfitDistance:              370 * priceAdjustment,
	TPDistanceShortForTighterSL:     200 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     40 * float64(priceAdjustment),
	TrendCandles:                    200,
	TrendDiff:                       220 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 27,
	PriceOffset:                     18 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             20 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
