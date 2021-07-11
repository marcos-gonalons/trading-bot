package types

// StrategyParams ...
type StrategyParams struct {
	RiskPercentage                  float64
	StopLossDistance                float32
	TakeProfitDistance              float32
	TPDistanceShortForBreakEvenSL   float64
	PriceOffset                     float64
	TrendCandles                    int
	TrendDiff                       float64
	CandlesAmountForHorizontalLevel int
	ValidTradingTimes               TradingTimes
}

type TradingTimes struct {
	ValidMonths    []string
	ValidWeekdays  []string
	ValidHalfHours []string
}
