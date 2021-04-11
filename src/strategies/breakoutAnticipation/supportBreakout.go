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

// SupportBreakoutStrategyName ...
const SupportBreakoutStrategyName = MainStrategyName + " - SBA"

func (s *Strategy) supportBreakoutAnticipationStrategy(candles []*types.Candle) {
	validMonths, validWeekdays, validHalfHours := getValidSupportBreakoutTimes()

	if !s.isExecutionTimeValid(validMonths, []string{}, []string{}) || !s.isExecutionTimeValid([]string{}, validWeekdays, []string{}) {
		s.log(SupportBreakoutStrategyName, "Today it's not the day for support breakout anticipation")
		return
	}

	isValidTimeToOpenAPosition := s.isExecutionTimeValid(
		validMonths,
		validWeekdays,
		validHalfHours,
	)

	if !isValidTimeToOpenAPosition {
		if len(s.positions) == 0 {
			s.savePendingOrder(ibroker.ShortSide)
		}
	} else {
		if s.pendingOrder != nil {
			s.createPendingOrder(ibroker.ShortSide)
			return
		}
		s.pendingOrder = nil
	}

	riskPercentage := float64(1)
	stopLossDistance := 15
	takeProfitDistance := 34
	candlesAmountWithLowerPriceToBeConsideredBottom := 14
	tpDistanceShortForBreakEvenSL := 1
	priceOffset := 2
	trendCandles := 90
	trendDiff := float64(30)

	if len(s.positions) > 0 {
		s.checkIfSLShouldBeMovedToBreakEven(float64(tpDistanceShortForBreakEvenSL), ibroker.ShortSide)
		return
	}

	lastCandlesIndex := len(candles) - 1
	price, err := s.HorizontalLevelsService.GetSupportPrice(candlesAmountWithLowerPriceToBeConsideredBottom)

	if err != nil {
		errorMessage := "Not a good short setup yet -> " + err.Error()
		s.log(SupportBreakoutStrategyName, errorMessage)
		return
	}

	price = price + float64(priceOffset)
	if price >= float64(s.currentBrokerQuote.Bid) {
		s.log(SupportBreakoutStrategyName, "Price is lower than the current ask, so we can't create the short order now. Price is -> "+utils.FloatToString(price, 2))
		s.log(SupportBreakoutStrategyName, "Quote is -> "+utils.GetStringRepresentation(s.currentBrokerQuote))
		return
	}

	// todo -> find a better name for closingOrdersTimestamp
	if s.closingOrdersTimestamp == candles[lastCandlesIndex].Timestamp {
		return
	}

	s.closingOrdersTimestamp = candles[lastCandlesIndex].Timestamp
	s.log(SupportBreakoutStrategyName, "Ok, we might have a short setup at price "+utils.FloatToString(price, 2))
	s.log(SupportBreakoutStrategyName, "Closing all working orders first if any ...")
	s.APIRetryFacade.CloseOrders(
		s.API.GetWorkingOrders(s.orders),
		retryFacade.RetryParams{
			DelayBetweenRetries: 5 * time.Second,
			MaxRetries:          30,
			SuccessCallback: func() {
				highestValue := candles[lastCandlesIndex-1].High
				for i := lastCandlesIndex - 1; i > lastCandlesIndex-trendCandles; i-- {
					if i < 1 {
						break
					}
					if candles[i].High > highestValue {
						highestValue = candles[i].High
					}
				}
				diff := highestValue - candles[lastCandlesIndex-1].High
				if diff < trendDiff {
					s.log(SupportBreakoutStrategyName, "At the end it wasn't a good short setup, doing nothing ...")
					return
				}

				s.APIRetryFacade.CloseOrders(
					s.API.GetWorkingOrders(s.orders),
					retryFacade.RetryParams{
						DelayBetweenRetries: 5 * time.Second,
						MaxRetries:          30,
						SuccessCallback: func() {
							float32Price := float32(price)

							stopLoss := float32Price + float32(stopLossDistance)
							takeProfit := float32Price - float32(takeProfitDistance)
							size := math.Floor((s.state.Equity*(riskPercentage/100))/float64(stopLossDistance+1) + 1)
							if size == 0 {
								size = 1
							}

							order := &api.Order{
								CurrentAsk: &s.currentBrokerQuote.Ask,
								CurrentBid: &s.currentBrokerQuote.Bid,
								Instrument: ibroker.GER30SymbolName,
								StopPrice:  &float32Price,
								Qty:        float32(size),
								Side:       ibroker.ShortSide,
								StopLoss:   &stopLoss,
								TakeProfit: &takeProfit,
								Type:       ibroker.StopType,
							}

							s.log(SupportBreakoutStrategyName, "Short order to be created -> "+utils.GetStringRepresentation(order))

							if !isValidTimeToOpenAPosition {
								s.log(SupportBreakoutStrategyName, "Now is not the time for opening any short orders, saving it for later ...")
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

func getValidSupportBreakoutTimes() ([]string, []string, []string) {
	validMonths := []string{"January", "March", "April", "May", "June", "August", "September", "October", "December"}
	validWeekdays := []string{"Monday", "Tuesday", "Thursday", "Friday"}
	validHalfHours := []string{"8:00", "8:30", "9:00", "10:00", "10:30", "11:00", "11:30", "12:00", "12:30", "13:00", "14:00", "14:30", "15:00", "15:30", "16:00", "16:30", "17:00", "18:00"}

	return validMonths, validWeekdays, validHalfHours
}
