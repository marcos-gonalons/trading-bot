package GBPUSD

import "TradingBot/src/types"

var priceAdjustment float32 = float32(1) / float32(10000)

var ResistanceBounceParams = types.TickerStrategyParams{
	RiskPercentage:                  5,
	StopLossDistance:                180 * priceAdjustment,
	TakeProfitDistance:              250 * priceAdjustment,
	TPDistanceShortForTighterSL:     50 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     50 * float64(priceAdjustment),
	TrendCandles:                    25,
	TrendDiff:                       130 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 15,
	PriceOffset:                     -10 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             21 * 24 * 60 * 60,
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
	StopLossDistance:                160 * priceAdjustment,
	TakeProfitDistance:              470 * priceAdjustment,
	TPDistanceShortForTighterSL:     0 * float64(priceAdjustment),
	SLDistanceWhenTPIsVeryClose:     0 * float64(priceAdjustment),
	TrendCandles:                    50,
	TrendDiff:                       10 * float64(priceAdjustment),
	CandlesAmountForHorizontalLevel: 52,
	PriceOffset:                     -4 * float64(priceAdjustment),
	MaxSecondsOpenTrade:             50 * 24 * 60 * 60,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{},
		ValidWeekdays:  []string{},
		ValidHalfHours: []string{},
	},
	MaxTradeExecutionPriceDifference: 9999,
	MinPositionSize:                  10000,
}
