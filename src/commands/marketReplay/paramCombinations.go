package main

import (
	"TradingBot/src/markets"
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/types"

	"github.com/thoas/go-funk"
)

type ParamCombinations struct {
	RiskPercentage []float64

	MinStopLossDistance []float64
	MaxStopLossDistance []float64
	StopLossDistance    []float64

	TakeProfitDistance []float64
	MinProfit          []float64

	TPDistanceShortForTighterSL []float64
	SLDistanceWhenTPIsVeryClose []float64

	SLDistanceShortForTighterTP []float64
	TPDistanceWhenSLIsVeryClose []float64

	FutureCandles []int
	PastCandles   []int

	CandlesAmountWithoutEMAsCrossing []int
	LimitAndStopOrderPriceOffset     []float64
	StopLossPriceOffset              []float64

	TrendCandles []int
	TrendDiff    []float64

	MaxTradeExecutionPriceDifference []float64
	MaxSecondsOpenTrade              []int64
	MinPositionSize                  []int64

	ValidTradingTimes []*types.TradingTimes

	WithPendingOrders     []bool
	CloseOrdersOnBadTrend []bool
}

func GetCombinations() *ParamCombinations {
	var c ParamCombinations

	var priceAdjustment float64 = float64(1) / float64(10000)

	c.RiskPercentage = []float64{1}
	c.MaxSecondsOpenTrade = []int64{0}
	c.MinPositionSize = []int64{10000}
	c.TrendCandles = []int{0}
	c.TrendDiff = funk.Map([]float64{1}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.MaxTradeExecutionPriceDifference = funk.Map([]float64{999999}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.LimitAndStopOrderPriceOffset = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.MinStopLossDistance = funk.Map([]float64{50}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.MaxStopLossDistance = funk.Map([]float64{600}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.StopLossDistance = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.TakeProfitDistance = funk.Map([]float64{200}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.MinProfit = funk.Map([]float64{999999}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.TPDistanceShortForTighterSL = funk.Map([]float64{30}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.SLDistanceWhenTPIsVeryClose = funk.Map([]float64{-90}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.SLDistanceShortForTighterTP = funk.Map([]float64{100}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.TPDistanceWhenSLIsVeryClose = funk.Map([]float64{-20}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.FutureCandles = []int{30}
	c.PastCandles = []int{40}
	c.CandlesAmountWithoutEMAsCrossing = []int{12}

	c.StopLossPriceOffset = funk.Map([]float64{75}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	return &c
}

func candlesLoopWithCombinations(
	csvLines [][]string,
	market markets.MarketInterface,
	container *services.Container,
	simulatorAPI api.Interface,
	c *ParamCombinations,
) {
	// todo: refactor
	for _, RiskPercentage := range c.RiskPercentage {
		for _, MinStopLossDistance := range c.MinStopLossDistance {
			for _, MaxStopLossDistance := range c.MaxStopLossDistance {
				for _, StopLossDistance := range c.StopLossDistance {
					for _, TakeProfitDistance := range c.TakeProfitDistance {
						for _, MinProfit := range c.MinProfit {
							for _, TPDistanceShortForTighterSL := range c.TPDistanceShortForTighterSL {
								for _, SLDistanceWhenTPIsVeryClose := range c.SLDistanceWhenTPIsVeryClose {
									for _, SLDistanceShortForTighterTP := range c.SLDistanceShortForTighterTP {
										for _, TPDistanceWhenSLIsVeryClose := range c.TPDistanceWhenSLIsVeryClose {
											for _, FutureCandles := range c.FutureCandles {
												for _, PastCandles := range c.PastCandles {
													for _, CandlesAmountWithoutEMAsCrossing := range c.CandlesAmountWithoutEMAsCrossing {
														for _, LimitAndStopOrderPriceOffset := range c.LimitAndStopOrderPriceOffset {
															for _, StopLossPriceOffset := range c.StopLossPriceOffset {
																for _, TrendCandles := range c.TrendCandles {
																	for _, TrendDiff := range c.TrendDiff {
																		for _, MaxTradeExecutionPriceDifference := range c.MaxTradeExecutionPriceDifference {
																			for _, MaxSecondsOpenTrade := range c.MaxSecondsOpenTrade {
																				for _, MinPositionSize := range c.MinPositionSize {
																					var longParams types.MarketStrategyParams
																					var shortParams types.MarketStrategyParams

																					longParams.RiskPercentage = RiskPercentage
																					longParams.MinStopLossDistance = float32(MinStopLossDistance)
																					longParams.MaxStopLossDistance = float32(MaxStopLossDistance)
																					longParams.StopLossDistance = float32(StopLossDistance)
																					longParams.TakeProfitDistance = float32(TakeProfitDistance)
																					longParams.MinProfit = float32(MinProfit)
																					longParams.TrailingStopLoss = &types.TrailingStopLoss{
																						TPDistanceShortForTighterSL: TPDistanceShortForTighterSL,
																						SLDistanceWhenTPIsVeryClose: SLDistanceWhenTPIsVeryClose,
																					}
																					longParams.TrailingTakeProfit = &types.TrailingTakeProfit{
																						SLDistanceShortForTighterTP: SLDistanceShortForTighterTP,
																						TPDistanceWhenSLIsVeryClose: TPDistanceWhenSLIsVeryClose,
																					}
																					longParams.CandlesAmountForHorizontalLevel = &types.CandlesAmountForHorizontalLevel{
																						Future: FutureCandles,
																						Past:   PastCandles,
																					}
																					longParams.CandlesAmountWithoutEMAsCrossing = CandlesAmountWithoutEMAsCrossing
																					longParams.LimitAndStopOrderPriceOffset = LimitAndStopOrderPriceOffset
																					longParams.StopLossPriceOffset = StopLossPriceOffset
																					longParams.TrendCandles = TrendCandles
																					longParams.TrendDiff = TrendDiff
																					longParams.MaxTradeExecutionPriceDifference = MaxTradeExecutionPriceDifference
																					longParams.MaxSecondsOpenTrade = MaxSecondsOpenTrade
																					longParams.MinPositionSize = MinPositionSize

																					market.SetStrategyParams(&longParams, &shortParams)

																					candlesLoop(csvLines, market, container, simulatorAPI)

																					// todo: print best
																				}
																			}
																		}
																	}
																}
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
