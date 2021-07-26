package GER30

import (
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"time"
)

func (s *Strategy) getResistanceBreakoutStrategyName() string {
	return s.BaseClass.Name + " - RBA"
}

func (s *Strategy) resistanceBreakoutAnticipationStrategy(candles []*types.Candle) {
	s.BaseClass.Log(s.getResistanceBreakoutStrategyName(), "resistanceBreakoutAnticipationStrategy started")
	defer func() {
		s.BaseClass.Log(s.getResistanceBreakoutStrategyName(), "resistanceBreakoutAnticipationStrategy ended")
	}()

	validMonths := ResistanceBreakoutParams.ValidTradingTimes.ValidMonths
	validWeekdays := ResistanceBreakoutParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := ResistanceBreakoutParams.ValidTradingTimes.ValidHalfHours

	if !s.isExecutionTimeValid(validMonths, []string{}, []string{}) || !s.isExecutionTimeValid([]string{}, validWeekdays, []string{}) {
		s.BaseClass.Log(s.getResistanceBreakoutStrategyName(), "Today it's not the day for resistance breakout anticipation for GER30")
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
		if s.BaseClass.GetPendingOrder() != nil {
			s.createPendingOrder(ibroker.LongSide)
		}
		s.BaseClass.SetPendingOrder(nil)
	}

	p := utils.FindPositionBySymbol(s.BaseClass.GetPositions(), s.BaseClass.GetSymbol().BrokerAPIName)
	if p != nil && p.Side == ibroker.LongSide {
		s.BaseClass.CheckIfSLShouldBeAdjusted(&ResistanceBreakoutParams, p)
	}

	lastCompletedCandleIndex := len(candles) - 2
	price, err := s.BaseClass.HorizontalLevelsService.GetResistancePrice(ResistanceBreakoutParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good long setup yet -> " + err.Error()
		s.BaseClass.Log(s.getResistanceBreakoutStrategyName(), errorMessage)
		return
	}

	price = price - ResistanceBreakoutParams.PriceOffset
	if price <= float64(s.BaseClass.GetCurrentBrokerQuote().Ask) {
		s.BaseClass.Log(s.getResistanceBreakoutStrategyName(), "Price is lower than the current ask, so we can't create the long order now. Price is -> "+utils.FloatToString(price, 2))
		s.BaseClass.Log(s.getResistanceBreakoutStrategyName(), "Quote is -> "+utils.GetStringRepresentation(s.BaseClass.GetCurrentBrokerQuote()))
		return
	}

	s.BaseClass.Log(s.getResistanceBreakoutStrategyName(), "Ok, we might have a long setup at price "+utils.FloatToString(price, 2))
	if !s.BaseClass.TrendsService.IsBullishTrend(
		ResistanceBreakoutParams.TrendCandles,
		ResistanceBreakoutParams.TrendDiff,
		candles,
		lastCompletedCandleIndex,
	) {
		s.BaseClass.Log(s.getResistanceBreakoutStrategyName(), "At the end it wasn't a good long setup, doing nothing ...")
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

	if utils.FindPositionBySymbol(s.BaseClass.GetPositions(), s.BaseClass.GetSymbol().BrokerAPIName) != nil {
		s.BaseClass.Log(s.getResistanceBreakoutStrategyName(), "There is an open position, no need to close any orders ...")
		s.onValidTradeSetup(params)
	} else {
		s.BaseClass.Log(s.getResistanceBreakoutStrategyName(), "There isn't any open position. Closing orders first ...")
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
