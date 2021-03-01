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

func (s *Strategy) resistanceBreakoutAnticipationStrategy(candles []*types.Candle) {
	validMonths, validWeekdays, validHalfHours := getValidResistanceBreakoutTimes()

	if !s.isExecutionTimeValid(validMonths, []string{}, []string{}) || !s.isExecutionTimeValid([]string{}, validWeekdays, []string{}) {
		s.Logger.Log("Today it's not the day for resistance breakout anticipation")
		return
	}

	isValidTimeToOpenAPosition := s.isExecutionTimeValid(
		validMonths,
		validWeekdays,
		validHalfHours,
	)

	if !isValidTimeToOpenAPosition {
		if len(s.positions) == 0 {
			s.savePendingOrder("buy")
		}
	} else {
		if s.pendingOrder != nil {
			s.createPendingOrder("buy")
			return
		}
		s.pendingOrder = nil
	}

	ignoreLastNCandles := 19
	riskPercentage := float64(1)
	stopLossDistance := 12
	takeProfitDistance := 27
	candlesAmountWithLowerPriceToBeConsideredTop := 19
	tpDistanceShortForBreakEvenSL := 5

	if len(s.positions) > 0 {
		s.checkIfSLShouldBeMovedToBreakEven(float64(tpDistanceShortForBreakEvenSL), "buy")
		return
	}

	lastCandlesIndex := len(candles) - 1
	for i := lastCandlesIndex - ignoreLastNCandles; i > lastCandlesIndex-ignoreLastNCandles-lastCandlesIndex; i-- {
		if i < 1 {
			break
		}
		isFalsePositive := false
		for j := i + 1; j < lastCandlesIndex-1; j++ {
			if candles[j].High >= candles[i].High {
				isFalsePositive = true
				break
			}
		}

		if isFalsePositive {
			s.Logger.Log("Doing nothing, not the proper trade setup for a buy 1")
			break
		}

		isFalsePositive = false
		for j := i - candlesAmountWithLowerPriceToBeConsideredTop; j < i; j++ {
			if j < 1 || j > lastCandlesIndex {
				continue
			}
			if candles[j].High >= candles[i].High {
				isFalsePositive = true
				break
			}
		}

		if isFalsePositive {
			s.Logger.Log("Doing nothing, not the proper trade setup for a buy 2")
			break
		}

		price := candles[i].High - 1
		if price <= float64(s.currentBrokerQuote.Ask) {
			s.Logger.Log("Price is lower than the current ask, so we can't create the long order now. Price is -> " + utils.FloatToString(price, 2))
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
			if order.Side == "buy" && order.ParentID == nil {
				ordersArr = append(ordersArr, order)
			}
		}
		s.Logger.Log("Ok, we might have a long setup, closing all working orders first if any ...")
		s.APIRetryFacade.CloseSpecificOrders(
			ordersArr,
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					lowestValue := candles[lastCandlesIndex-1].Low
					for i := lastCandlesIndex - 1; i > lastCandlesIndex-180; i-- {
						if i < 1 {
							break
						}
						if candles[i].Low < lowestValue {
							lowestValue = candles[i].Low
						}
					}
					diff := candles[lastCandlesIndex-1].Low - lowestValue
					if diff < 10 {
						s.Logger.Log("At the end it wasn't a good setup, doing nothing ...")
						return
					}

					s.APIRetryFacade.CloseAllWorkingOrders(
						retryFacade.RetryParams{
							DelayBetweenRetries: 5 * time.Second,
							MaxRetries:          30,
							SuccessCallback: func() {
								float32Price := float32(price)

								stopLoss := float32Price - float32(stopLossDistance)
								takeProfit := float32Price + float32(takeProfitDistance)
								size := math.Floor((s.state.Equity*(riskPercentage/100))/float64(stopLossDistance+2) + 1)
								if size == 0 {
									size = 1
								}

								order := &api.Order{
									Instrument: ibroker.GER30SymbolName,
									StopPrice:  &float32Price,
									Qty:        float32(size),
									Side:       "buy",
									StopLoss:   &stopLoss,
									TakeProfit: &takeProfit,
									Type:       "stop",
								}

								s.Logger.Log("Buy order to be created -> " + utils.GetStringRepresentation(order))

								if !isValidTimeToOpenAPosition {
									s.Logger.Log("Now is not the time for opening any buy orders, saving it for later ...")
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

func getValidResistanceBreakoutTimes() ([]string, []string, []string) {
	validMonths := []string{"January", "March", "April", "May", "June", "August", "September", "October"}
	validWeekdays := []string{"Monday", "Tuesday", "Wednesday", "Friday"}
	validHalfHours := []string{"9:00", "10:00", "11:30", "12:00", "12:30", "20:30"}

	return validMonths, validWeekdays, validHalfHours
}
