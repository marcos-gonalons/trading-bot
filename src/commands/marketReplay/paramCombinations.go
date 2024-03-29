package main

import (
	"TradingBot/src/manager"
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/simulator"
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
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
	RangesMinimumDistanceToLevel                   []float64
	RangesPriceOffset                              []float64
	RangesRangePoints                              []int
	RangesStartWith                                []types.LevelType
	RangesTakeProfitStrategy                       []string
	RangesStopLossStrategy                         []string
	RangesOrderType                                []string
	RangesTrendyOnly                               []bool
}

type BestCombination struct {
	BestProfits     float64
	BestCombination *types.MarketStrategyParams
}

const REPORT_FILE_PATH = "./.combinations-report.txt"

var successRateThreshold float64 = float64(40) / float64(100)
var parallelRoutines = 20

var mutex *sync.Mutex = &sync.Mutex{}
var completed int64 = 0

func GetCombinations(minPositionSize int64) (*ParamCombinations, int64) {
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

	c.FutureCandles = []int{35}
	c.PastCandles = []int{35}

	c.MaxStopLossDistance = funk.Map([]float64{100}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.TakeProfitDistance = funk.Map([]float64{0, 20, 40, 60, 80, 100}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.StopLossDistance = funk.Map([]float64{-80, -50, -20, 0, 20, 40, 60, 80}, func(r float64) float64 { return r * priceAdjustment }).([]float64)

	c.RangesCandlesToCheck = []int64{400}
	c.RangesMaxPriceDifferenceForSameHorizontalLevel = funk.Map([]float64{400}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.RangesMinPriceDifferenceBetweenRangePoints = funk.Map([]float64{300}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.RangesMinCandlesBetweenRangePoints = []int64{5}
	c.RangesMaxCandlesBetweenRangePoints = []int64{500}
	c.RangesMinimumDistanceToLevel = funk.Map([]float64{150}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.RangesPriceOffset = funk.Map([]float64{0}, func(r float64) float64 { return r * priceAdjustment }).([]float64)
	c.RangesRangePoints = []int{3}
	c.RangesStartWith = []types.LevelType{types.SUPPORT_TYPE}
	c.RangesTakeProfitStrategy = []string{"level-with-offset"}
	c.RangesStopLossStrategy = []string{"level-with-offset"}
	c.RangesOrderType = []string{constants.MarketType}
	c.RangesTrendyOnly = []bool{false}

	return &c, getTotalLength(&c)
}

func candlesLoopWithCombinations(
	csvLines [][]string,
	marketName string,
	c *ParamCombinations,
	combinationsLength int64,
	side Side,
	strategy string,
) {
	bestCombination := &BestCombination{
		BestProfits:     -999999,
		BestCombination: nil,
	}
	iteration := int64(0)

	createReportFile()

	fmt.Println("Combinations length -> " + strconv.FormatInt(combinationsLength, 10))

	var waitingGroup sync.WaitGroup
	waitingGroup.Add(parallelRoutines)

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
																										for _, RangesMinimumDistanceToLevel := range c.RangesMinimumDistanceToLevel {
																											for _, RangesPriceOffset := range c.RangesPriceOffset {
																												for _, RangesRangePoints := range c.RangesRangePoints {
																													for _, RangesStartWith := range c.RangesStartWith {
																														for _, RangesTakeProfitStrategy := range c.RangesTakeProfitStrategy {
																															for _, RangesStopLossStrategy := range c.RangesStopLossStrategy {
																																for _, RangesOrderType := range c.RangesOrderType {
																																	for _, RangesTrendyOnly := range c.RangesTrendyOnly {
																																		iteration++

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
																																		params.Ranges.MinimumDistanceToLevel = RangesMinimumDistanceToLevel
																																		params.Ranges.PriceOffset = RangesPriceOffset
																																		params.Ranges.RangePoints = RangesRangePoints
																																		params.Ranges.StartWith = RangesStartWith
																																		params.Ranges.TakeProfitStrategy = RangesTakeProfitStrategy
																																		params.Ranges.StopLossStrategy = RangesStopLossStrategy
																																		params.Ranges.OrderType = RangesOrderType
																																		params.Ranges.TrendyOnly = RangesTrendyOnly

																																		params.PositionSizeStrategy = positionSize.BASED_ON_MIN_SIZE

																																		go func(p types.MarketStrategyParams) {
																																			executeCombination(
																																				params,
																																				bestCombination,
																																				marketName,
																																				side,
																																				csvLines,
																																				combinationsLength,
																																				&waitingGroup,
																																				strategy,
																																			)
																																		}(params)

																																		if iteration%int64(parallelRoutines) == 0 {
																																			waitingGroup.Wait()

																																			if iteration == combinationsLength {
																																				break
																																			}

																																			waitingGroup = sync.WaitGroup{}

																																			if combinationsLength-iteration < int64(parallelRoutines) {
																																				waitingGroup.Add(int(combinationsLength - iteration))
																																			} else {
																																				waitingGroup.Add(parallelRoutines)
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
			}
		}
	}

	waitingGroup.Wait()

	write("\n\n\nDone! Best combination -> " + utils.FloatToString(bestCombination.BestProfits, 2))
	write(utils.GetStringRepresentation(bestCombination))
}

func getTotalLength(c *ParamCombinations) (length int64) {
	v := reflect.ValueOf(*c)

	length = int64(1)
	for i := 0; i < v.NumField(); i++ {
		l := v.Field(i).Len()
		if l > 0 {
			length *= int64(l)
		}
	}

	return
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

func executeCombination(
	params types.MarketStrategyParams,
	bestCombination *BestCombination,
	marketName string,
	side Side,
	csvLines [][]string,
	combinationsLength int64,
	waitingGroup *sync.WaitGroup,
	strategy string,
) {
	defer waitingGroup.Done()

	start := time.Now().UnixMilli()

	container := services.Container{}
	container.Initialize(false)

	simulatorAPI := simulator.CreateAPIServiceInstance(
		&api.Credentials{},
		container.HttpClient,
		container.Logger,
	)

	container.SetAPI(simulatorAPI)

	manager := &manager.Manager{
		ServicesContainer: &container,
	}

	market := getMarketInstance(manager, marketName)

	simulatorAPI.CloseAllOrders()
	simulatorAPI.CloseAllPositions()
	simulatorAPI.SetTrades(nil)

	longsParamsFunc, shortsParamsFunc := getSetStrategyParamsFuncs(strategy, market)

	if side == LONGS_SIDE {
		longsParamsFunc(&params)
	} else {
		shortsParamsFunc(&params)
	}
	candlesLoop(csvLines, market, simulatorAPI, false)

	state, _ := simulatorAPI.GetState()
	profits := state.Balance - initialBalance

	trades := simulatorAPI.GetTrades()
	profitableTrades := float64(0)
	for _, trade := range trades {
		if trade.Result > 0 {
			profitableTrades++
		}
	}
	successRate := profitableTrades / float64(len(trades))

	if profits > bestCombination.BestProfits && successRate >= float64(successRateThreshold) {
		mutex.Lock()
		bestCombination.BestProfits = profits
		bestCombination.BestCombination = &params

		write("\n\nNew best combination " + utils.FloatToString(profits, 2))
		write("\nTotal trades " + strconv.Itoa(len(trades)))
		write("\nProfitable trades " + utils.FloatToString(float64(profitableTrades), 0))
		write("\nSuccess rate " + utils.FloatToString(float64(successRate*100), 2))
		write("\n" + utils.GetStringRepresentation(bestCombination))
		mutex.Unlock()
	}

	mutex.Lock()
	completed++
	mutex.Unlock()

	combinationTime := (time.Now().UnixMilli() - start) / int64(parallelRoutines)
	combinationsLeft := combinationsLength - completed
	estimatedRemainingMilliseconds := combinationTime * int64(combinationsLeft)

	progress := fmt.Sprintf(""+
		"Progress: %.4f%% | Estimated remaining time: %s seconds",
		float64(completed)*100.0/float64(combinationsLength),
		strconv.Itoa(int(estimatedRemainingMilliseconds)/1000),
	)
	fmt.Println(progress)
}
