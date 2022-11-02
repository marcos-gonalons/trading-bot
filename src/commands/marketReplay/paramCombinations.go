package main

import (
	"TradingBot/src/markets"
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"fmt"

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

	ValidTradingTimes []*types.TradingTimes

	WithPendingOrders     []bool
	CloseOrdersOnBadTrend []bool
}

func GetCombinations() *ParamCombinations {
	return nil
	var c ParamCombinations

	var priceAdjustment float64 = float64(1) / float64(10000)

	c.RiskPercentage = []float64{1}
	c.MaxSecondsOpenTrade = []int64{0}
	c.TrendCandles = []int{0}
	c.TrendDiff = funk.Map([]float64{1}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.MaxTradeExecutionPriceDifference = funk.Map([]float64{999999}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.LimitAndStopOrderPriceOffset = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.StopLossPriceOffset = funk.Map([]float64{25}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.MinStopLossDistance = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.MaxStopLossDistance = funk.Map([]float64{620, 640, 660, 680, 700}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.StopLossDistance = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.TakeProfitDistance = funk.Map([]float64{330}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.MinProfit = funk.Map([]float64{220}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.TPDistanceShortForTighterSL = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.SLDistanceWhenTPIsVeryClose = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.SLDistanceShortForTighterTP = funk.Map([]float64{40}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.TPDistanceWhenSLIsVeryClose = funk.Map([]float64{-180}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.FutureCandles = []int{25}
	c.PastCandles = []int{45}
	c.CandlesAmountWithoutEMAsCrossing = []int{21}

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
	state, _ := simulatorAPI.GetState()

	var initialBalance = state.Balance
	var bestProfits float64 = -999999
	var bestCombination *types.MarketStrategyParams

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
																				var params types.MarketStrategyParams

																				params.RiskPercentage = RiskPercentage
																				params.MinStopLossDistance = MinStopLossDistance
																				params.MaxStopLossDistance = MaxStopLossDistance
																				params.StopLossDistance = StopLossDistance
																				params.TakeProfitDistance = TakeProfitDistance
																				params.MinProfit = MinProfit
																				params.TrailingStopLoss = &types.TrailingStopLoss{
																					TPDistanceShortForTighterSL: TPDistanceShortForTighterSL,
																					SLDistanceWhenTPIsVeryClose: SLDistanceWhenTPIsVeryClose,
																				}
																				params.TrailingTakeProfit = &types.TrailingTakeProfit{
																					SLDistanceShortForTighterTP: SLDistanceShortForTighterTP,
																					TPDistanceWhenSLIsVeryClose: TPDistanceWhenSLIsVeryClose,
																				}
																				params.CandlesAmountForHorizontalLevel = &types.CandlesAmountForHorizontalLevel{
																					Future: FutureCandles,
																					Past:   PastCandles,
																				}
																				params.CandlesAmountWithoutEMAsCrossing = CandlesAmountWithoutEMAsCrossing
																				params.LimitAndStopOrderPriceOffset = LimitAndStopOrderPriceOffset
																				params.StopLossPriceOffset = StopLossPriceOffset
																				params.TrendCandles = TrendCandles
																				params.TrendDiff = TrendDiff
																				params.MaxTradeExecutionPriceDifference = MaxTradeExecutionPriceDifference
																				params.MaxSecondsOpenTrade = MaxSecondsOpenTrade

																				market.GetCandlesHandler().SetCandles([]*types.Candle{})

																				simulatorAPI.CloseAllOrders()
																				simulatorAPI.CloseAllPositions()
																				simulatorAPI.SetTrades(0)
																				simulatorAPI.SetState(&api.State{
																					Balance:      initialBalance,
																					UnrealizedPL: 0,
																					Equity:       initialBalance,
																				})

																				state, _ := simulatorAPI.GetState()

																				//market.SetStrategyParams(&params, nil)
																				market.SetStrategyParams(nil, &params)
																				candlesLoop(csvLines, market, container, simulatorAPI)

																				state, _ = simulatorAPI.GetState()
																				profits := state.Balance - initialBalance

																				if profits > bestProfits {
																					bestProfits = profits
																					bestCombination = &params

																					fmt.Println("New best combination", profits, utils.GetStringRepresentation(bestCombination))
																					fmt.Println("\n\n")
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

	fmt.Println("\n\n\nDone! Best combination -> ", bestProfits, utils.GetStringRepresentation(bestCombination))
}
