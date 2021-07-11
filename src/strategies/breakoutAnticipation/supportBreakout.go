package breakoutAnticipation

import (
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"time"
)

// SupportBreakoutStrategyName ...
const SupportBreakoutStrategyName = MainStrategyName + " - SBA"

func (s *Strategy) supportBreakoutAnticipationStrategy(candles []*types.Candle) {
	s.log(SupportBreakoutStrategyName, "supportBreakoutAnticipationStrategy started")
	defer func() {
		s.log(SupportBreakoutStrategyName, "supportBreakoutAnticipationStrategy ended")
	}()

	validMonths := SupportBreakoutParams.ValidTradingTimes.ValidMonths
	validWeekdays := SupportBreakoutParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := SupportBreakoutParams.ValidTradingTimes.ValidHalfHours

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
		s.savePendingOrder(ibroker.ShortSide)
	} else {
		if s.pendingOrder != nil {
			s.createPendingOrder(ibroker.ShortSide)
		}
		s.pendingOrder = nil
	}

	p := s.getOpenPosition()
	if p != nil && p.Side == ibroker.ShortSide {
		s.checkIfSLShouldBeMovedToBreakEven(SupportBreakoutParams.TPDistanceShortForBreakEvenSL, p)
	}

	lastCompletedCandleIndex := len(candles) - 2
	price, err := s.HorizontalLevelsService.GetSupportPrice(SupportBreakoutParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good short setup yet -> " + err.Error()
		s.log(SupportBreakoutStrategyName, errorMessage)
		return
	}

	price = price + SupportBreakoutParams.PriceOffset
	if price >= float64(s.currentBrokerQuote.Bid) {
		s.log(SupportBreakoutStrategyName, "Price is lower than the current ask, so we can't create the short order now. Price is -> "+utils.FloatToString(price, 2))
		s.log(SupportBreakoutStrategyName, "Quote is -> "+utils.GetStringRepresentation(s.currentBrokerQuote))
		return
	}

	s.log(SupportBreakoutStrategyName, "Ok, we might have a short setup at price "+utils.FloatToString(price, 2))
	highestValue := candles[lastCompletedCandleIndex].High
	for i := lastCompletedCandleIndex; i > lastCompletedCandleIndex-SupportBreakoutParams.TrendCandles; i-- {
		if i < 1 {
			break
		}
		if candles[i].High > highestValue {
			highestValue = candles[i].High
		}
	}
	diff := highestValue - candles[lastCompletedCandleIndex].High
	if diff < SupportBreakoutParams.TrendDiff {
		s.log(SupportBreakoutStrategyName, "At the end it wasn't a good short setup")
		if s.getOpenPosition() == nil {
			s.log(SupportBreakoutStrategyName, "There isn't an open position, closing short orders ...")
			s.APIRetryFacade.CloseOrders(
				s.API.GetWorkingOrderWithBracketOrders(ibroker.ShortSide, s.GetSymbol().BrokerAPIName, s.orders),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback:     func() { s.orders = nil },
				},
			)
		}
		return
	}

	params := OnValidTradeSetupParams{
		Price:              price,
		StopLossDistance:   SupportBreakoutParams.StopLossDistance,
		TakeProfitDistance: SupportBreakoutParams.TakeProfitDistance,
		RiskPercentage:     SupportBreakoutParams.RiskPercentage,
		IsValidTime:        isValidTimeToOpenAPosition,
		Side:               ibroker.ShortSide,
	}

	if s.getOpenPosition() != nil {
		s.log(SupportBreakoutStrategyName, "There is an open position, no need to close any orders ...")
		s.onValidTradeSetup(params)
	} else {
		s.log(SupportBreakoutStrategyName, "There isn't any open position. Closing orders first ...")
		s.APIRetryFacade.CloseOrders(
			s.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.orders, s.GetSymbol().BrokerAPIName)),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					s.onValidTradeSetup(params)
				},
			},
		)
	}

}
