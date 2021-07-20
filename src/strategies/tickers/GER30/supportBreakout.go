package GER30

import (
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"time"
)

func (s *Strategy) getSupportBreakoutStrategyName() string {
	return s.BaseClass.Name + " - SBA"
}

func (s *Strategy) supportBreakoutAnticipationStrategy(candles []*types.Candle) {
	s.BaseClass.Log(s.getSupportBreakoutStrategyName(), "supportBreakoutAnticipationStrategy started")
	defer func() {
		s.BaseClass.Log(s.getSupportBreakoutStrategyName(), "supportBreakoutAnticipationStrategy ended")
	}()

	validMonths := SupportBreakoutParams.ValidTradingTimes.ValidMonths
	validWeekdays := SupportBreakoutParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := SupportBreakoutParams.ValidTradingTimes.ValidHalfHours

	if !s.isExecutionTimeValid(validMonths, []string{}, []string{}) || !s.isExecutionTimeValid([]string{}, validWeekdays, []string{}) {
		s.BaseClass.Log(s.getSupportBreakoutStrategyName(), "Today it's not the day for support breakout anticipation for GER30")
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

	p := utils.FindPositionBySymbol(s.BaseClass.GetPositions(), s.BaseClass.GetSymbol().BrokerAPIName)
	if p != nil && p.Side == ibroker.ShortSide {
		s.checkIfSLShouldBeMovedToBreakEven(SupportBreakoutParams.TPDistanceShortForTighterSL, p)
	}

	lastCompletedCandleIndex := len(candles) - 2
	price, err := s.BaseClass.HorizontalLevelsService.GetSupportPrice(SupportBreakoutParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good short setup yet -> " + err.Error()
		s.BaseClass.Log(s.getSupportBreakoutStrategyName(), errorMessage)
		return
	}

	price = price + SupportBreakoutParams.PriceOffset
	if price >= float64(s.BaseClass.GetCurrentBrokerQuote().Bid) {
		s.BaseClass.Log(s.getSupportBreakoutStrategyName(), "Price is lower than the current ask, so we can't create the short order now. Price is -> "+utils.FloatToString(price, 2))
		s.BaseClass.Log(s.getSupportBreakoutStrategyName(), "Quote is -> "+utils.GetStringRepresentation(s.BaseClass.GetCurrentBrokerQuote()))
		return
	}

	s.BaseClass.Log(s.getSupportBreakoutStrategyName(), "Ok, we might have a short setup at price "+utils.FloatToString(price, 2))
	if !s.BaseClass.TrendsService.IsBearishTrend(
		SupportBreakoutParams.TrendCandles,
		SupportBreakoutParams.TrendDiff,
		candles,
		lastCompletedCandleIndex,
	) {
		s.BaseClass.Log(s.getSupportBreakoutStrategyName(), "At the end it wasn't a good short setup")
		if utils.FindPositionBySymbol(s.BaseClass.GetPositions(), s.BaseClass.GetSymbol().BrokerAPIName) == nil {
			s.BaseClass.Log(s.getSupportBreakoutStrategyName(), "There isn't an open position, closing short orders ...")
			s.BaseClass.APIRetryFacade.CloseOrders(
				s.BaseClass.API.GetWorkingOrderWithBracketOrders(ibroker.ShortSide, s.BaseClass.GetSymbol().BrokerAPIName, s.BaseClass.GetOrders()),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback:     func() { s.BaseClass.SetOrders(nil) },
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

	if utils.FindPositionBySymbol(s.BaseClass.GetPositions(), s.BaseClass.GetSymbol().BrokerAPIName) != nil {
		s.BaseClass.Log(s.getSupportBreakoutStrategyName(), "There is an open position, no need to close any orders ...")
		s.onValidTradeSetup(params)
	} else {
		s.BaseClass.Log(s.getSupportBreakoutStrategyName(), "There isn't any open position. Closing orders first ...")
		s.BaseClass.APIRetryFacade.CloseOrders(
			s.BaseClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseClass.GetOrders(), s.BaseClass.GetSymbol().BrokerAPIName)),
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
