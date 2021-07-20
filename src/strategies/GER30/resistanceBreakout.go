package GER30

import (
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"time"
)

/**
	todo: separated class for resistance and support strategies?
	Will have the main parent strategy class as dependency?
**/

// ResistanceBreakoutStrategyName ...
const ResistanceBreakoutStrategyName = MainStrategyName + " - RBA"

func (s *Strategy) resistanceBreakoutAnticipationStrategy(candles []*types.Candle) {
	s.log(ResistanceBreakoutStrategyName, "resistanceBreakoutAnticipationStrategy started")
	defer func() {
		s.log(ResistanceBreakoutStrategyName, "resistanceBreakoutAnticipationStrategy ended")
	}()

	validMonths := ResistanceBreakoutParams.ValidTradingTimes.ValidMonths
	validWeekdays := ResistanceBreakoutParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := ResistanceBreakoutParams.ValidTradingTimes.ValidHalfHours

	if !s.isExecutionTimeValid(validMonths, []string{}, []string{}) || !s.isExecutionTimeValid([]string{}, validWeekdays, []string{}) {
		s.log(ResistanceBreakoutStrategyName, "Today it's not the day for resistance breakout anticipation for GER30")
		return
	}

	isValidTimeToOpenAPosition := s.isExecutionTimeValid(
		validMonths,
		validWeekdays,
		validHalfHours,
	)

	if !isValidTimeToOpenAPosition {
		s.savePendingOrder(ibroker.LongSide)
	} else {
		if s.pendingOrder != nil {
			s.createPendingOrder(ibroker.LongSide)
		}
		s.pendingOrder = nil
	}

	p := s.getOpenPosition()
	if p != nil && p.Side == ibroker.LongSide {
		s.checkIfSLShouldBeMovedToBreakEven(ResistanceBreakoutParams.TPDistanceShortForBreakEvenSL, p)
	}

	lastCompletedCandleIndex := len(candles) - 2
	price, err := s.HorizontalLevelsService.GetResistancePrice(ResistanceBreakoutParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good long setup yet -> " + err.Error()
		s.log(ResistanceBreakoutStrategyName, errorMessage)
		return
	}

	price = price - ResistanceBreakoutParams.PriceOffset
	if price <= float64(s.currentBrokerQuote.Ask) {
		s.log(ResistanceBreakoutStrategyName, "Price is lower than the current ask, so we can't create the long order now. Price is -> "+utils.FloatToString(price, 2))
		s.log(ResistanceBreakoutStrategyName, "Quote is -> "+utils.GetStringRepresentation(s.currentBrokerQuote))
		return
	}

	s.log(ResistanceBreakoutStrategyName, "Ok, we might have a long setup at price "+utils.FloatToString(price, 2))
	if !s.TrendsService.IsBullishTrend(
		ResistanceBreakoutParams.TrendCandles,
		ResistanceBreakoutParams.TrendDiff,
		candles,
		lastCompletedCandleIndex,
	) {
		s.log(ResistanceBreakoutStrategyName, "At the end it wasn't a good long setup, doing nothing ...")
		return
	}

	params := OnValidTradeSetupParams{
		Price:              price,
		StopLossDistance:   ResistanceBreakoutParams.StopLossDistance,
		TakeProfitDistance: ResistanceBreakoutParams.TakeProfitDistance,
		RiskPercentage:     ResistanceBreakoutParams.RiskPercentage,
		IsValidTime:        isValidTimeToOpenAPosition,
		Side:               ibroker.LongSide,
	}

	if s.getOpenPosition() != nil {
		s.log(ResistanceBreakoutStrategyName, "There is an open position, no need to close any orders ...")
		s.onValidTradeSetup(params)
	} else {
		s.log(ResistanceBreakoutStrategyName, "There isn't any open position. Closing orders first ...")
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
