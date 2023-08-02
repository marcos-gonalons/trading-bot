package main

import (
	"TradingBot/src/markets"
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

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
	MaxAttemptsToGetSL               []int

	TrendCandles []int
	TrendDiff    []float64

	MaxTradeExecutionPriceDifference []float64
	MaxSecondsOpenTrade              []int64

	ValidTradingTimes []*types.TradingTimes

	WithPendingOrders     []bool
	CloseOrdersOnBadTrend []bool

	RangesCandlesToCheck                           []int64
	RangesMaxPriceDifferenceForSameHorizontalLevel []float64
	RangesMinPriceDifferenceBetweenRangePoints     []float64
	RangesMinCandlesBetweenRangePoints             []int64
	RangesMaxCandlesBetweenRangePoints             []int64
	RangesPriceOffset                              []float64
	RangesRangePoints                              []int
	RangesStartWith                                []types.LevelType
	RangesTakeProfitStrategy                       []string
	RangesStopLossStrategy                         []string
	RangesOrderType                                []string
	RangesTrendyOnly                               []bool
}

const REPORT_FILE_PATH = "./.combinations-report.txt"

func GetCombinations(minPositionSize int64) (*ParamCombinations, int) {
	var c ParamCombinations

	var priceAdjustment float64 = float64(1) / float64(minPositionSize)

	c.RiskPercentage = []float64{1}
	/** not used for ranges */
	c.MinStopLossDistance = []float64{0}
	/** not used for ranges */
	c.MinProfit = []float64{0}
	/** not used for ranges */
	c.CandlesAmountWithoutEMAsCrossing = []int{0}
	/** not used for ranges */
	c.LimitAndStopOrderPriceOffset = []float64{0}
	/** not used for ranges */
	c.StopLossPriceOffset = []float64{0}
	/** not used for ranges */
	c.MaxAttemptsToGetSL = []int{0}
	/** not used for ranges */
	c.TrendCandles = []int{0}
	/** not used for ranges */
	c.TrendDiff = []float64{0}

	c.MaxSecondsOpenTrade = []int64{0}
	c.MaxTradeExecutionPriceDifference = funk.Map([]float64{999999}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.TPDistanceShortForTighterSL = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.SLDistanceWhenTPIsVeryClose = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.SLDistanceShortForTighterTP = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.TPDistanceWhenSLIsVeryClose = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.FutureCandles = []int{3, 10, 20}
	c.PastCandles = []int{3, 10, 20}

	c.MaxStopLossDistance = funk.Map([]float64{300}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.TakeProfitDistance = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.StopLossDistance = funk.Map([]float64{20, 50, 80, 120, 150}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.RangesCandlesToCheck = []int64{400}
	c.RangesMaxPriceDifferenceForSameHorizontalLevel = funk.Map([]float64{75, 50, 25}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.RangesMinPriceDifferenceBetweenRangePoints = funk.Map([]float64{25, 50, 75, 100, 140}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.RangesMinCandlesBetweenRangePoints = []int64{5}
	c.RangesMaxCandlesBetweenRangePoints = []int64{300}
	c.RangesPriceOffset = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.RangesRangePoints = []int{3}
	c.RangesStartWith = []types.LevelType{"resistance"}
	c.RangesTakeProfitStrategy = []string{"half"}
	c.RangesStopLossStrategy = []string{"distance"}
	c.RangesOrderType = []string{constants.LimitType}
	c.RangesTrendyOnly = []bool{true}

	return &c, getTotalLength(&c)
}

func candlesLoopWithCombinations(
	csvLines [][]string,
	market markets.MarketInterface,
	container *services.Container,
	simulatorAPI api.Interface,
	c *ParamCombinations,
	combinationsLength int,
	side Side,
	longsParamsFunc func(p *types.MarketStrategyParams),
	shortsParamsFunc func(p *types.MarketStrategyParams),
) {
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
																for _, MaxAttemptsToGetSL := range c.MaxAttemptsToGetSL {
																	for _, TrendCandles := range c.TrendCandles {
																		for _, TrendDiff := range c.TrendDiff {
																			for _, MaxTradeExecutionPriceDifference := range c.MaxTradeExecutionPriceDifference {
																				for _, MaxSecondsOpenTrade := range c.MaxSecondsOpenTrade {
																					for _, RangesCandlesToCheck := range c.RangesCandlesToCheck {
																						for _, RangesMaxPriceDifferenceForSameHorizontalLevel := range c.RangesMaxPriceDifferenceForSameHorizontalLevel {
																							for _, RangesMinPriceDifferenceBetweenRangePoints := range c.RangesMinPriceDifferenceBetweenRangePoints {
																								for _, RangesMinCandlesBetweenRangePoints := range c.RangesMinCandlesBetweenRangePoints {
																									for _, RangesMaxCandlesBetweenRangePoints := range c.RangesMaxCandlesBetweenRangePoints {
																										for _, RangesPriceOffset := range c.RangesPriceOffset {
																											for _, RangesRangePoints := range c.RangesRangePoints {
																												for _, RangesStartWith := range c.RangesStartWith {
																													for _, RangesTakeProfitStrategy := range c.RangesTakeProfitStrategy {
																														for _, RangesStopLossStrategy := range c.RangesStopLossStrategy {
																															for _, RangesOrderType := range c.RangesOrderType {
																																for _, RangesTrendyOnly := range c.RangesTrendyOnly {

																																	start := time.Now().UnixMilli()

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
																																	params.EmaCrossover.CandlesAmountWithoutEMAsCrossing = CandlesAmountWithoutEMAsCrossing
																																	params.LimitAndStopOrderPriceOffset = LimitAndStopOrderPriceOffset
																																	params.EmaCrossover.StopLossPriceOffset = StopLossPriceOffset
																																	params.EmaCrossover.MaxAttemptsToGetSL = MaxAttemptsToGetSL
																																	params.TrendCandles = TrendCandles
																																	params.TrendDiff = TrendDiff
																																	params.MaxTradeExecutionPriceDifference = MaxTradeExecutionPriceDifference
																																	params.MaxSecondsOpenTrade = MaxSecondsOpenTrade

																																	params.Ranges.CandlesToCheck = RangesCandlesToCheck
																																	params.Ranges.MaxPriceDifferenceForSameHorizontalLevel = RangesMaxPriceDifferenceForSameHorizontalLevel
																																	params.Ranges.MinPriceDifferenceBetweenRangePoints = RangesMinPriceDifferenceBetweenRangePoints
																																	params.Ranges.MinCandlesBetweenRangePoints = RangesMinCandlesBetweenRangePoints
																																	params.Ranges.MaxCandlesBetweenRangePoints = RangesMaxCandlesBetweenRangePoints
																																	params.Ranges.PriceOffset = RangesPriceOffset
																																	params.Ranges.RangePoints = RangesRangePoints
																																	params.Ranges.StartWith = RangesStartWith
																																	params.Ranges.TakeProfitStrategy = RangesTakeProfitStrategy
																																	params.Ranges.StopLossStrategy = RangesStopLossStrategy
																																	params.Ranges.OrderType = RangesOrderType
																																	params.Ranges.TrendyOnly = RangesTrendyOnly

																																	params.PositionSizeStrategy = positionSize.BASED_ON_MIN_SIZE

																																	market.GetCandlesHandler().SetCandles([]*types.Candle{})

																																	simulatorAPI.CloseAllOrders()
																																	simulatorAPI.CloseAllPositions()
																																	simulatorAPI.SetTrades(nil)

																																	if side == LONGS_SIDE {
																																		longsParamsFunc(&params)
																																	} else {
																																		shortsParamsFunc(&params)
																																	}
																																	candlesLoop(csvLines, market, container, simulatorAPI, false)

																																	state, _ := simulatorAPI.GetState()
																																	profits := state.Balance - initialBalance

																																	if profits > bestProfits {
																																		bestProfits = profits
																																		bestCombination = &params

																																		write("\n\nNew best combination " + utils.FloatToString(profits, 2))
																																		write("\nTotal trades " + strconv.Itoa(len(simulatorAPI.GetTrades())))
																																		write("\n" + utils.GetStringRepresentation(bestCombination))
																																	}

																																	i++

																																	combinationTime := time.Now().UnixMilli() - start
																																	combinationsLeft := combinationsLength - i
																																	estimatedRemainingMilliseconds := combinationTime * int64(combinationsLeft)

																																	progress := fmt.Sprintf(""+
																																		"Progress: %.4f%% | Estimated remaining time: %s seconds",
																																		float64(i)*100.0/float64(combinationsLength),
																																		strconv.Itoa(int(estimatedRemainingMilliseconds)/1000),
																																	)
																																	fmt.Println(progress)
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

	v = strings.Replace(v, ",", "\n", -1)

	file.Write([]byte(v))
}
