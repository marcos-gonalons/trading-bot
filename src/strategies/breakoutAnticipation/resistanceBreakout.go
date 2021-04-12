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

/**
	todo: separated class for resistance and support strategies?
	Will have the main parent strategy class as dependency?
**/

// ResistanceBreakoutStrategyName ...
const ResistanceBreakoutStrategyName = MainStrategyName + " - RBA"

func (s *Strategy) resistanceBreakoutAnticipationStrategy(candles []*types.Candle) {
	validMonths, validWeekdays, validHalfHours := getValidResistanceBreakoutTimes()

	if !s.isExecutionTimeValid(validMonths, []string{}, []string{}) || !s.isExecutionTimeValid([]string{}, validWeekdays, []string{}) {
		s.log(ResistanceBreakoutStrategyName, "Today it's not the day for resistance breakout anticipation")
		return
	}

	isValidTimeToOpenAPosition := s.isExecutionTimeValid(
		validMonths,
		validWeekdays,
		validHalfHours,
	)

	if !isValidTimeToOpenAPosition {
		if len(s.positions) == 0 {
			s.savePendingOrder(ibroker.LongSide)
		}
	} else {
		if s.pendingOrder != nil {
			s.createPendingOrder(ibroker.LongSide)
			return
		}
		s.pendingOrder = nil
	}

	riskPercentage := float64(1)
	stopLossDistance := 16
	takeProfitDistance := 34
	candlesAmountWithLowerPriceToBeConsideredTop := 21
	tpDistanceShortForBreakEvenSL := 2
	priceOffset := 1
	trendCandles := 90
	trendDiff := float64(5)

	if len(s.positions) > 0 {
		s.checkIfSLShouldBeMovedToBreakEven(float64(tpDistanceShortForBreakEvenSL), ibroker.LongSide)
		return
	}

	lastCompletedCandleIndex := len(candles) - 2
	price, err := s.HorizontalLevelsService.GetResistancePrice(candlesAmountWithLowerPriceToBeConsideredTop, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good long setup yet -> " + err.Error()
		s.log(ResistanceBreakoutStrategyName, errorMessage)
		return
	}

	price = price - float64(priceOffset)
	if price <= float64(s.currentBrokerQuote.Ask) {
		s.log(ResistanceBreakoutStrategyName, "Price is lower than the current ask, so we can't create the long order now. Price is -> "+utils.FloatToString(price, 2))
		s.log(ResistanceBreakoutStrategyName, "Quote is -> "+utils.GetStringRepresentation(s.currentBrokerQuote))
		return
	}

	// todo -> find a better name for closingOrdersTimestamp
	if s.closingOrdersTimestamp == candles[lastCompletedCandleIndex].Timestamp {
		return
	}

	s.closingOrdersTimestamp = candles[lastCompletedCandleIndex].Timestamp
	s.log(ResistanceBreakoutStrategyName, "Ok, we might have a long setup at price "+utils.FloatToString(price, 2))
	s.log(ResistanceBreakoutStrategyName, "Closing all working orders first if any ...")
	s.APIRetryFacade.CloseOrders(
		s.API.GetWorkingOrders(s.orders),
		retryFacade.RetryParams{
			DelayBetweenRetries: 5 * time.Second,
			MaxRetries:          30,
			SuccessCallback: func() {
				lowestValue := candles[lastCompletedCandleIndex].Low
				for i := lastCompletedCandleIndex; i > lastCompletedCandleIndex-trendCandles; i-- {
					if i < 1 {
						break
					}
					if candles[i].Low < lowestValue {
						lowestValue = candles[i].Low
					}
				}
				diff := candles[lastCompletedCandleIndex].Low - lowestValue
				if diff < trendDiff {
					s.log(ResistanceBreakoutStrategyName, "At the end it wasn't a good setup, doing nothing ...")
					return
				}

				// todo: find out why i am closing the orders twice
				// maybe it's not needed here

				s.APIRetryFacade.CloseOrders(
					s.API.GetWorkingOrders(s.orders),
					retryFacade.RetryParams{
						DelayBetweenRetries: 5 * time.Second,
						MaxRetries:          30,
						SuccessCallback: func() {
							float32Price := float32(price)

							stopLoss := float32Price - float32(stopLossDistance)
							takeProfit := float32Price + float32(takeProfitDistance)
							size := math.Floor((s.state.Equity*(riskPercentage/100))/float64(stopLossDistance+1) + 1)
							if size == 0 {
								size = 1
							}

							order := &api.Order{
								Instrument: ibroker.GER30SymbolName,
								StopPrice:  &float32Price,
								Qty:        float32(size),
								Side:       ibroker.LongSide,
								StopLoss:   &stopLoss,
								TakeProfit: &takeProfit,
								Type:       ibroker.StopType,
							}

							s.log(ResistanceBreakoutStrategyName, "Buy order to be created -> "+utils.GetStringRepresentation(order))

							if !isValidTimeToOpenAPosition {
								s.log(ResistanceBreakoutStrategyName, "Now is not the time for opening any buy orders, saving it for later ...")
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

func getValidResistanceBreakoutTimes() ([]string, []string, []string) {
	validMonths := []string{"January", "February", "March", "April", "May", "June"}
	validWeekdays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	validHalfHours := []string{"9:00", "10:00", "10:30", "11:00", "11:30", "12:00", "12:30", "15:30", "16:00", "16:30", "17:00", "18:30", "19:00", "19:30", "20:00", "20:30", "21:00"}

	return validMonths, validWeekdays, validHalfHours
}
