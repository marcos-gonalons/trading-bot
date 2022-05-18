package types

// MarketStrategyParams ...
type MarketStrategyParams struct {
	RiskPercentage                   float64
	StopLossDistance                 float32
	TakeProfitDistance               float32
	TPDistanceShortForTighterSL      float64
	SLDistanceWhenTPIsVeryClose      float64
	PriceOffset                      float64
	TrendCandles                     int
	TrendDiff                        float64
	CandlesAmountForHorizontalLevel  int
	ValidTradingTimes                TradingTimes
	MaxTradeExecutionPriceDifference float64
	MaxSecondsOpenTrade              int64
	MinPositionSize                  int64
}

type TradingTimes struct {
	ValidMonths    []string
	ValidWeekdays  []string
	ValidHalfHours []string
}
