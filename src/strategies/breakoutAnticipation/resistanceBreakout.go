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
	s.log(ResistanceBreakoutStrategyName, "resistanceBreakoutAnticipationStrategy started")
	defer func() {
		s.log(ResistanceBreakoutStrategyName, "resistanceBreakoutAnticipationStrategy ended")
	}()

	validMonths := s.GetSymbol().ValidTradingTimes.Longs.ValidMonths
	validWeekdays := s.GetSymbol().ValidTradingTimes.Longs.ValidWeekdays
	validHalfHours := s.GetSymbol().ValidTradingTimes.Longs.ValidHalfHours

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
		s.savePendingOrder(ibroker.LongSide)
	} else {
		if s.pendingOrder != nil {
			s.createPendingOrder(ibroker.LongSide)
		}
		s.pendingOrder = nil
	}

	riskPercentage := float64(1.5)
	stopLossDistance := float32(24)
	takeProfitDistance := float32(34)
	candlesAmountWithLowerPriceToBeConsideredTop := 24
	tpDistanceShortForBreakEvenSL := 0
	priceOffset := 1
	trendCandles := 60
	trendDiff := float64(15)

	p := s.getOpenPosition()
	if p != nil && p.Side == ibroker.LongSide {
		s.checkIfSLShouldBeMovedToBreakEven(float64(tpDistanceShortForBreakEvenSL), p)
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

	s.log(ResistanceBreakoutStrategyName, "Ok, we might have a long setup at price "+utils.FloatToString(price, 2))
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
		s.log(ResistanceBreakoutStrategyName, "At the end it wasn't a good long setup, doing nothing ...")
		return
	}

	params := CreateLongOrderParams{
		Price:              price,
		StopLossDistance:   stopLossDistance,
		TakeProfitDistance: takeProfitDistance,
		RiskPercentage:     riskPercentage,
		IsValidTime:        isValidTimeToOpenAPosition,
	}

	if s.getOpenPosition() != nil {
		s.log(ResistanceBreakoutStrategyName, "There is an open position, no need to close any orders ...")
		s.createLongOrder(params)
	} else {
		s.log(ResistanceBreakoutStrategyName, "There isn't any open position. Closing orders first ...")
		s.APIRetryFacade.CloseOrders(
			s.API.GetWorkingOrders(s.orders),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					s.createLongOrder(params)
				},
			},
		)
	}

}

// TODO: refactor this, since this method is the same as createShortOrder
type CreateLongOrderParams struct {
	Price              float64
	StopLossDistance   float32
	TakeProfitDistance float32
	RiskPercentage     float64
	IsValidTime        bool
}

func (s *Strategy) createLongOrder(params CreateLongOrderParams) {
	float32Price := float32(params.Price)

	stopLoss := float32Price - float32(params.StopLossDistance)
	takeProfit := float32Price + float32(params.TakeProfitDistance)
	size := math.Floor((s.state.Equity*(params.RiskPercentage/100))/float64(params.StopLossDistance+1) + 1)
	if size == 0 {
		size = 1
	}

	order := &api.Order{
		Instrument: s.GetSymbol().BrokerAPIName,
		StopPrice:  &float32Price,
		Qty:        float32(size),
		Side:       ibroker.LongSide,
		StopLoss:   &stopLoss,
		TakeProfit: &takeProfit,
		Type:       ibroker.StopType,
	}

	s.log(ResistanceBreakoutStrategyName, "Buy order to be created -> "+utils.GetStringRepresentation(order))

	if s.getOpenPosition() != nil {
		s.log(ResistanceBreakoutStrategyName, "There is an open position, saving the order for later ...")
		s.pendingOrder = order
		return
	}

	if !params.IsValidTime {
		s.log(ResistanceBreakoutStrategyName, "Now is not the time for opening any buy orders, saving it for later ...")
		s.pendingOrder = order
		return
	}

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
