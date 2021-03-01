package breakoutAnticipation

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"math"
	"time"
)

func (s *Strategy) supportBreakoutAnticipationStrategy(candles []*types.Candle) {
	validMonths, validWeekdays, validHalfHours := getValidSupportBreakoutTimes()

	if !s.isExecutionTimeValid(validMonths, []string{}, []string{}) || !s.isExecutionTimeValid([]string{}, validWeekdays, []string{}) {
		s.Logger.Log("Today it's not the day for support breakout anticipation")
		return
	}

	isValidTimeToOpenAPosition := s.isExecutionTimeValid(
		validMonths,
		validWeekdays,
		validHalfHours,
	)

	if !isValidTimeToOpenAPosition {
		if len(s.positions) == 0 {
			s.savePendingOrder("sell")
		}
	} else {
		if s.pendingOrder != nil {
			s.createPendingOrder("sell")
			return
		}
		s.pendingOrder = nil
	}

	ignoreLastNCandles := 15
	riskPercentage := float64(1)
	stopLossDistance := 12
	takeProfitDistance := 27
	candlesAmountWithLowerPriceToBeConsideredBottom := 15
	tpDistanceShortForBreakEvenSL := 5

	if len(s.positions) > 0 {
		s.checkIfSLShouldBeMovedToBreakEven(float64(tpDistanceShortForBreakEvenSL), "sell")
		return
	}

	lastCandlesIndex := len(candles) - 1
	for i := lastCandlesIndex - ignoreLastNCandles; i > lastCandlesIndex-ignoreLastNCandles-lastCandlesIndex; i-- {
		if i < 1 {
			break
		}
		isFalsePositive := false
		for j := i + 1; j < lastCandlesIndex-1; j++ {
			if candles[j].Low <= candles[i].Low {
				isFalsePositive = true
				break
			}
		}

		if isFalsePositive {
			s.Logger.Log("Doing nothing, not the proper trade setup for a short 1")
			break
		}

		isFalsePositive = false
		for j := i - candlesAmountWithLowerPriceToBeConsideredBottom; j < i; j++ {
			if j < 1 || j > lastCandlesIndex {
				continue
			}
			if candles[j].Low <= candles[i].Low {
				isFalsePositive = true
				break
			}
		}

		if isFalsePositive {
			s.Logger.Log("Doing nothing, not the proper trade setup for a short 2")
			break
		}

		price := candles[i].Low + 3
		if price >= float64(s.currentBrokerQuote.Bid) {
			s.Logger.Log("Price is lower than the current ask, so we can't create the short order now. Price is -> " + utils.FloatToString(price, 2))
			s.Logger.Log("Quote is -> " + utils.GetStringRepresentation(s.currentBrokerQuote))
			continue
		}

		// todo -> find a better name for closingOrdersTimestamp
		if s.closingOrdersTimestamp == candles[lastCandlesIndex].Timestamp {
			return
		}

		s.closingOrdersTimestamp = candles[lastCandlesIndex].Timestamp
		workingOrders := utils.GetWorkingOrders(s.orders)
		var ordersArr []*api.Order
		for _, order := range workingOrders {
			if order.Side == "sell" && order.ParentID == nil {
				ordersArr = append(ordersArr, order)
			}
		}
		s.Logger.Log("Ok, we might have a short setup, closing all working orders first if any ...")
		s.APIRetryFacade.CloseSpecificOrders(
			ordersArr,
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					highestValue := candles[lastCandlesIndex-1].High
					for i := lastCandlesIndex - 1; i > lastCandlesIndex-120; i-- {
						if i < 1 {
							break
						}
						if candles[i].High > highestValue {
							highestValue = candles[i].High
						}
					}
					diff := highestValue - candles[lastCandlesIndex-1].High
					if diff < 29 {
						s.Logger.Log("At the end it wasn't a good short setup, doing nothing ...")
						return
					}

					s.APIRetryFacade.CloseAllWorkingOrders(
						retryFacade.RetryParams{
							DelayBetweenRetries: 5 * time.Second,
							MaxRetries:          30,
							SuccessCallback: func() {
								float32Price := float32(price)

								stopLoss := float32Price + float32(stopLossDistance)
								takeProfit := float32Price - float32(takeProfitDistance)
								size := math.Floor((s.state.Equity*(riskPercentage/100))/float64(stopLossDistance+2) + 1)
								if size == 0 {
									size = 1
								}

								order := &api.Order{
									CurrentAsk: &s.currentBrokerQuote.Ask,
									CurrentBid: &s.currentBrokerQuote.Bid,
									Instrument: ibroker.GER30SymbolName,
									StopPrice:  &float32Price,
									Qty:        float32(size),
									Side:       "sell",
									StopLoss:   &stopLoss,
									TakeProfit: &takeProfit,
									Type:       "stop",
								}

								s.Logger.Log("Short order to be created -> " + utils.GetStringRepresentation(order))

								if !isValidTimeToOpenAPosition {
									s.Logger.Log("Now is not the time for opening any short orders, saving it for later ...")
									s.pendingOrder = order
								} else {
									s.APIRetryFacade.CreateOrder(
										order,
										func() *api.Quote {
											return s.currentBrokerQuote
										},
										s.setStringValues,
										retryFacade.RetryParams{
											DelayBetweenRetries: 10 * time.Second,
											MaxRetries:          20,
										},
									)
								}
							},
						},
					)
				},
			},
		)
	}
}

func getValidSupportBreakoutTimes() ([]string, []string, []string) {
	validMonths := []string{"March", "April", "June", "September", "December"}
	validWeekdays := []string{"Monday", "Tuesday", "Thursday", "Friday"}
	validHalfHours := []string{"8:30", "9:00", "12:00", "13:00", "14:30", "15:30", "18:00"}

	return validMonths, validWeekdays, validHalfHours
}
