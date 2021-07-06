package breakoutAnticipation

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
}

var ResistanceBreakoutParams = StrategyParams{
	RiskPercentage:                  1.5,
	StopLossDistance:                24,
	TakeProfitDistance:              34,
	CandlesAmountForHorizontalLevel: 24,
	TPDistanceShortForBreakEvenSL:   0,
	PriceOffset:                     1,
	TrendCandles:                    60,
	TrendDiff:                       15,
}

var SupportBreakoutParams = StrategyParams{
	RiskPercentage:                  1,
	StopLossDistance:                15,
	TakeProfitDistance:              34,
	CandlesAmountForHorizontalLevel: 14,
	TPDistanceShortForBreakEvenSL:   1,
	PriceOffset:                     2,
	TrendCandles:                    90,
	TrendDiff:                       30,
}
