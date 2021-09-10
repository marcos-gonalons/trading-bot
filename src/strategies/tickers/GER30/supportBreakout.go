package GER30

import (
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"time"
)

func (s *Strategy) getSupportBreakoutStrategyName() string {
	return s.BaseTickerClass.Name + " - SBA"
}

func (s *Strategy) supportBreakoutAnticipationStrategy(candles []*types.Candle) {
	s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), "supportBreakoutAnticipationStrategy started")
	defer func() {
		s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), "supportBreakoutAnticipationStrategy ended")
	}()

	validMonths := SupportBreakoutParams.ValidTradingTimes.ValidMonths
	validWeekdays := SupportBreakoutParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := SupportBreakoutParams.ValidTradingTimes.ValidHalfHours

	if !s.isExecutionTimeValid(validMonths, []string{}, []string{}) || !s.isExecutionTimeValid([]string{}, validWeekdays, []string{}) {
		s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), "Today it's not the day for support breakout anticipation for GER30")
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
		if s.BaseTickerClass.GetPendingOrder() != nil {
			s.createPendingOrder(ibroker.ShortSide)
		}
		s.BaseTickerClass.SetPendingOrder(nil)
	}

	p := utils.FindPositionBySymbol(s.BaseTickerClass.GetPositions(), s.BaseTickerClass.GetSymbol().BrokerAPIName)
	if p != nil && p.Side == ibroker.ShortSide {
		s.BaseTickerClass.CheckIfSLShouldBeAdjusted(&SupportBreakoutParams, p)
	}

	lastCompletedCandleIndex := len(candles) - 2
	price, err := s.BaseTickerClass.HorizontalLevelsService.GetSupportPrice(SupportBreakoutParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good short setup yet -> " + err.Error()
		s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), errorMessage)
		return
	}

	price = price + SupportBreakoutParams.PriceOffset
	if price >= float64(s.BaseTickerClass.GetCurrentBrokerQuote().Bid) {
		s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), "Price is lower than the current ask, so we can't create the short order now. Price is -> "+utils.FloatToString(price, 2))
		s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), "Quote is -> "+utils.GetStringRepresentation(s.BaseTickerClass.GetCurrentBrokerQuote()))
		return
	}

	s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), "Ok, we might have a short setup at price "+utils.FloatToString(price, 2))
	if !s.BaseTickerClass.TrendsService.IsBearishTrend(
		SupportBreakoutParams.TrendCandles,
		SupportBreakoutParams.TrendDiff,
		candles,
		lastCompletedCandleIndex,
	) {
		s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), "At the end it wasn't a good short setup")
		if utils.FindPositionBySymbol(s.BaseTickerClass.GetPositions(), s.BaseTickerClass.GetSymbol().BrokerAPIName) == nil {
			s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), "There isn't an open position, closing short orders ...")
			s.BaseTickerClass.APIRetryFacade.CloseOrders(
				s.BaseTickerClass.API.GetWorkingOrderWithBracketOrders(ibroker.ShortSide, s.BaseTickerClass.GetSymbol().BrokerAPIName, s.BaseTickerClass.GetOrders()),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback:     func() { s.BaseTickerClass.SetOrders(nil) },
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

	if utils.FindPositionBySymbol(s.BaseTickerClass.GetPositions(), s.BaseTickerClass.GetSymbol().BrokerAPIName) != nil {
		s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), "There is an open position, no need to close any orders ...")
		s.onValidTradeSetup(params)
	} else {
		s.BaseTickerClass.Log(s.getSupportBreakoutStrategyName(), "There isn't any open position. Closing orders first ...")
		s.BaseTickerClass.APIRetryFacade.CloseOrders(
			s.BaseTickerClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseTickerClass.GetOrders(), s.BaseTickerClass.GetSymbol().BrokerAPIName)),
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
