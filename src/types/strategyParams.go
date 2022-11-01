package types

// MarketStrategyParams ...
// todo: document meaning of each param
// group params that are only used for a specific strategy
type MarketStrategyParams struct {
	RiskPercentage float64

	MinStopLossDistance float64
	MaxStopLossDistance float64
	StopLossDistance    float64

	TakeProfitDistance float64
	MinProfit          float64

	TrailingStopLoss   *TrailingStopLoss
	TrailingTakeProfit *TrailingTakeProfit

	CandlesAmountForHorizontalLevel  *CandlesAmountForHorizontalLevel
	CandlesAmountWithoutEMAsCrossing int
	LimitAndStopOrderPriceOffset     float64
	StopLossPriceOffset              float64

	TrendCandles int
	TrendDiff    float64

	MaxTradeExecutionPriceDifference float64
	MaxSecondsOpenTrade              int64
	MinPositionSize                  int64

	ValidTradingTimes *TradingTimes

	WithPendingOrders     bool
	CloseOrdersOnBadTrend bool
}

type TradingTimes struct {
	ValidMonths    []string
	ValidWeekdays  []string
	ValidHalfHours []string
}

type CandlesAmountForHorizontalLevel struct {
	Future int
	Past   int
}

type TrailingStopLoss struct {
	TPDistanceShortForTighterSL float64
	SLDistanceWhenTPIsVeryClose float64
}

type TrailingTakeProfit struct {
	SLDistanceShortForTighterTP float64
	TPDistanceWhenSLIsVeryClose float64
}
