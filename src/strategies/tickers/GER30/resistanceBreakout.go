package GER30

import (
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"time"
)

func (s *Strategy) getResistanceBreakoutStrategyName() string {
	return s.BaseTickerClass.Name + " - RBA"
}

func (s *Strategy) resistanceBreakoutAnticipationStrategy(candles []*types.Candle) {
	s.BaseTickerClass.Log(s.getResistanceBreakoutStrategyName(), "resistanceBreakoutAnticipationStrategy started")
	defer func() {
		s.BaseTickerClass.Log(s.getResistanceBreakoutStrategyName(), "resistanceBreakoutAnticipationStrategy ended")
	}()

	validMonths := ResistanceBreakoutParams.ValidTradingTimes.ValidMonths
	validWeekdays := ResistanceBreakoutParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := ResistanceBreakoutParams.ValidTradingTimes.ValidHalfHours

	if !s.isExecutionTimeValid(validMonths, []string{}, []string{}) || !s.isExecutionTimeValid([]string{}, validWeekdays, []string{}) {
		s.BaseTickerClass.Log(s.getResistanceBreakoutStrategyName(), "Today it's not the day for resistance breakout anticipation for GER30")
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
		if s.BaseTickerClass.GetPendingOrder() != nil {
			s.createPendingOrder(ibroker.LongSide)
		}
		s.BaseTickerClass.SetPendingOrder(nil)
	}

	p := utils.FindPositionBySymbol(s.BaseTickerClass.GetPositions(), s.BaseTickerClass.GetSymbol().BrokerAPIName)
	if p != nil && p.Side == ibroker.LongSide {
		s.BaseTickerClass.CheckIfSLShouldBeAdjusted(&ResistanceBreakoutParams, p)
	}

	lastCompletedCandleIndex := len(candles) - 2
	price, err := s.BaseTickerClass.HorizontalLevelsService.GetResistancePrice(ResistanceBreakoutParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good long setup yet -> " + err.Error()
		s.BaseTickerClass.Log(s.getResistanceBreakoutStrategyName(), errorMessage)
		return
	}

	price = price - ResistanceBreakoutParams.PriceOffset
	if price <= float64(s.BaseTickerClass.GetCurrentBrokerQuote().Ask) {
		s.BaseTickerClass.Log(s.getResistanceBreakoutStrategyName(), "Price is lower than the current ask, so we can't create the long order now. Price is -> "+utils.FloatToString(price, 2))
		s.BaseTickerClass.Log(s.getResistanceBreakoutStrategyName(), "Quote is -> "+utils.GetStringRepresentation(s.BaseTickerClass.GetCurrentBrokerQuote()))
		return
	}

	s.BaseTickerClass.Log(s.getResistanceBreakoutStrategyName(), "Ok, we might have a long setup at price "+utils.FloatToString(price, 2))
	if !s.BaseTickerClass.TrendsService.IsBullishTrend(
		ResistanceBreakoutParams.TrendCandles,
		ResistanceBreakoutParams.TrendDiff,
		candles,
		lastCompletedCandleIndex,
	) {
		s.BaseTickerClass.Log(s.getResistanceBreakoutStrategyName(), "At the end it wasn't a good long setup, doing nothing ...")
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

	if utils.FindPositionBySymbol(s.BaseTickerClass.GetPositions(), s.BaseTickerClass.GetSymbol().BrokerAPIName) != nil {
		s.BaseTickerClass.Log(s.getResistanceBreakoutStrategyName(), "There is an open position, no need to close any orders ...")
		s.onValidTradeSetup(params)
	} else {
		s.BaseTickerClass.Log(s.getResistanceBreakoutStrategyName(), "There isn't any open position. Closing orders first ...")
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
