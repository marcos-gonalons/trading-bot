package main

import (
	"TradingBot/src/markets"
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"fmt"
	"os"
	"reflect"

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

const LONGS_OR_SHORTS = "longs"
const REPORT_FILE_PATH = "./.combinations-report.txt"

func GetCombinations() (*ParamCombinations, int) {
	var c ParamCombinations

	var priceAdjustment float64 = float64(1) / float64(10000)

	c.RiskPercentage = []float64{1}
	c.MaxSecondsOpenTrade = []int64{0}
	c.TrendCandles = []int{0}
	c.TrendDiff = funk.Map([]float64{1}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.MaxTradeExecutionPriceDifference = funk.Map([]float64{999999}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.LimitAndStopOrderPriceOffset = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.StopLossDistance = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.StopLossPriceOffset = funk.Map([]float64{-150, -125, -100, -75, -50, -25, 0, 25, 50, 75, 100, 125, 150}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.MinStopLossDistance = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.MaxStopLossDistance = funk.Map([]float64{800}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.TakeProfitDistance = funk.Map([]float64{20, 50, 80, 110, 140, 170, 200, 230, 260, 290, 320, 350, 380, 410}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.MinProfit = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.TPDistanceShortForTighterSL = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.SLDistanceWhenTPIsVeryClose = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.SLDistanceShortForTighterTP = funk.Map([]float64{40, 3}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.TPDistanceWhenSLIsVeryClose = funk.Map([]float64{-180}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.FutureCandles = []int{0, 10, 25, 40}
	c.PastCandles = []int{0, 10, 25, 40}
	c.CandlesAmountWithoutEMAsCrossing = []int{0}

	return &c, getTotalLength(&c)
}

func candlesLoopWithCombinations(
	csvLines [][]string,
	market markets.MarketInterface,
	container *services.Container,
	simulatorAPI api.Interface,
	c *ParamCombinations,
	combinationsLength int,
) {
	// todo: refactor
	state, _ := simulatorAPI.GetState()

	var initialBalance = state.Balance
	var bestProfits float64 = -999999
	var bestCombination *types.MarketStrategyParams

	i := 0

	createReportFile()

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

																				if LONGS_OR_SHORTS == "longs" {
																					market.SetStrategyParams(&params, nil)
																				} else {
																					market.SetStrategyParams(nil, &params)
																				}
																				candlesLoop(csvLines, market, container, simulatorAPI)

																				state, _ = simulatorAPI.GetState()
																				profits := state.Balance - initialBalance

																				if profits > bestProfits {
																					bestProfits = profits
																					bestCombination = &params

																					write("\n\nNew best combination " + utils.FloatToString(profits, 2))
																					write("\n" + utils.GetStringRepresentation(bestCombination))
																				}

																				fmt.Println(float64(i)*100.0/float64(combinationsLength), "%")
																				i++
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

	write("\n\n\nDone! Best combination -> " + utils.FloatToString(bestProfits, 2))
	write(utils.GetStringRepresentation(bestCombination))
}

func getTotalLength(c *ParamCombinations) int {
	v := reflect.ValueOf(*c)

	length := 1
	for i := 0; i < v.NumField(); i++ {
		l := v.Field(i).Len()
		if l > 0 {
			length *= l
		}
	}

	return length
}

func createReportFile() {
	os.Remove(REPORT_FILE_PATH)

	file, err := os.OpenFile(REPORT_FILE_PATH, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		panic("error creating report file" + err.Error())
	}
	defer file.Close()
	if err != nil {
		file, err = os.Create(REPORT_FILE_PATH)
		if err != nil {
			panic("error creating report file")
		}
		defer file.Close()
	}
}

func write(v string) {
	fmt.Println(v)

	file, err := os.OpenFile(REPORT_FILE_PATH, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		panic("error writing to file" + err.Error())
	}
	defer file.Close()

	file.Write([]byte(v))
}
