package types

import (
	"TradingBot/src/services/positionSize"
)

// todo: document meaning of each param
type MarketStrategyParams struct {
	RiskPercentage float64

	MinStopLossDistance float64
	MaxStopLossDistance float64
	StopLossDistance    float64

	TakeProfitDistance float64
	MinProfit          float64

	TrailingStopLoss   *TrailingStopLoss
	TrailingTakeProfit *TrailingTakeProfit

	CandlesAmountForHorizontalLevel *CandlesAmountForHorizontalLevel
	LimitAndStopOrderPriceOffset    float64

	TrendCandles int
	TrendDiff    float64

	MaxTradeExecutionPriceDifference float64
	MaxSecondsOpenTrade              int64

	ValidTradingTimes *TradingTimes

	WithPendingOrders     bool
	CloseOrdersOnBadTrend bool

	PositionSizeStrategy positionSize.Strategy

	EmaCrossover EmaCrossover
	Ranges       Ranges
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

type EmaCrossover struct {
	StopLossPriceOffset              float64
	MaxAttemptsToGetSL               int
	CandlesAmountWithoutEMAsCrossing int
}

type LevelType string

const (
	RESISTANCE_TYPE LevelType = "resistance"
	SUPPORT_TYPE    LevelType = "support"
)

type Ranges struct {
	CandlesToCheck                           int64
	MaxPriceDifferenceForSameHorizontalLevel float64
	MinPriceDifferenceBetweenRangePoints     float64
	MinCandlesBetweenRangePoints             int64
	MaxCandlesBetweenRangePoints             int64
	MinimumDistanceToLevel                   float64
	RangePoints                              int
	PriceOffset                              float64
	StartWith                                LevelType
	TakeProfitStrategy                       string // "level" | "half" | "levelWithOffset" | "distance";
	StopLossStrategy                         string // "level" | "half" | "levelWithOffset" | "distance";
	OrderType                                string // refactor, use OrderType custom type (limit,stop,market)
	TrendyOnly                               bool
}
