package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/utils"
	"math"
	"strconv"
	"time"
)

func (s *Strategy) breakoutAnticipationStrategy() {
	if s.isCurrentTimeOutsideTradingHours() {
		s.Logger.Log("Doing nothing - Now it's not the time.")
		s.closeAllWorkingOrders(
			func() {
				s.orders = nil
				s.pendingOrder = nil
				s.closePositions(func() { s.positions = nil }, func(err error) {})
			},
		)

		return
	}

	if s.averageSpread > 3 {
		s.Logger.Log("Doing nothing since the spread is very big -> " + utils.FloatToString(s.averageSpread, 0))
		s.pendingOrder = nil
		s.closeAllWorkingOrders(func() { s.orders = nil })
		return
	}

	if len(s.candles) < 2 {
		return
	}

	s.resistanceBreakoutAnticipationStrategy()
	s.supportBreakoutAnticipationStrategy()
}

func (s *Strategy) resistanceBreakoutAnticipationStrategy() {
	validMonths := []string{"January", "March", "April", "May", "June", "August", "September", "October"}
	validWeekdays := []string{"Monday", "Tuesday", "Wednesday", "Friday"}
	validHalfHours := []string{"9:00", "10:00", "11:30", "12:00", "12:30", "20:30"}

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

	ignoreLastNCandles := 18
	riskPercentage := float64(1)
	stopLossDistance := 12
	takeProfitDistance := 27
	candlesAmountWithLowerPriceToBeConsideredTop := 18
	tpDistanceShortForBreakEvenSL := 5

	if len(s.positions) > 0 {
		s.checkIfSLShouldBeMovedToBreakEven(float64(tpDistanceShortForBreakEvenSL), "buy")
		return
	}

	lastCandlesIndex := len(s.candles) - 1
	for i := lastCandlesIndex - ignoreLastNCandles; i > lastCandlesIndex-ignoreLastNCandles-lastCandlesIndex; i-- {
		if i < 1 {
			break
		}
		isFalsePositive := false
		for j := i + 1; j < lastCandlesIndex-1; j++ {
			if s.candles[j].High >= s.candles[i].High {
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
			if s.candles[j].High >= s.candles[i].High {
				isFalsePositive = true
				break
			}
		}

		if isFalsePositive {
			s.Logger.Log("Doing nothing, not the proper trade setup for a buy 2")
			break
		}

		price := s.candles[i].High - 1
		if price <= float64(s.currentBrokerQuote.Ask) {
			s.Logger.Log("Price is lower than the current ask, so we can't create the long order now. Price is -> " + utils.FloatToString(price, 2))
			s.Logger.Log("Quote is -> " + utils.GetStringRepresentation(s.currentBrokerQuote))
			continue
		}

		// todo -> find a better name for closingOrdersTimestamp
		if s.closingOrdersTimestamp == s.candles[lastCandlesIndex].Timestamp {
			return
		}

		s.closingOrdersTimestamp = s.candles[lastCandlesIndex].Timestamp
		workingOrders := utils.GetWorkingOrders(s.orders)
		var ordersArr []*api.Order
		for _, order := range workingOrders {
			if order.Side == "buy" && order.ParentID == nil {
				ordersArr = append(ordersArr, order)
			}
		}
		s.Logger.Log("Ok, we might have a long setup, closing all working orders first if any ...")
		s.closeSpecificOrders(
			ordersArr,
			func() {
				lowestValue := s.candles[lastCandlesIndex-1].Low
				for i := lastCandlesIndex - 1; i > lastCandlesIndex-180; i-- {
					if i < 1 {
						break
					}
					if s.candles[i].Low < lowestValue {
						lowestValue = s.candles[i].Low
					}
				}
				diff := s.candles[lastCandlesIndex-1].Low - lowestValue
				if diff < 10 {
					s.Logger.Log("At the end it wasn't a good setup, doing nothing ...")
					return
				}

				s.closeAllWorkingOrders(
					func() {
						float32Price := float32(price)

						stopLoss := float32Price - float32(stopLossDistance)
						takeProfit := float32Price + float32(takeProfitDistance)
						size := math.Floor((s.state.Equity*(riskPercentage/100))/float64(stopLossDistance) + 1)
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
							s.createOrder(order, 20, 10*time.Second)
						}
					},
				)
			},
		)
	}

}

func (s *Strategy) supportBreakoutAnticipationStrategy() {
	validMonths := []string{"March", "April", "June", "September", "December"}
	validWeekdays := []string{"Monday", "Tuesday", "Thursday", "Friday"}
	validHalfHours := []string{"8:30", "9:00", "12:00", "13:00", "14:30", "15:30", "18:00"}

	if !s.isExecutionTimeValid(validMonths, []string{}, []string{}) || !s.isExecutionTimeValid([]string{}, validWeekdays, []string{}) {
		s.Logger.Log("Today it's not the day for support breakout anticipation")
		return
	}

	isValidTimeToOpenAPosition := s.isExecutionTimeValid(
		validMonths,
		validWeekdays,
		validHalfHours,
	)

	if !isValidTimeToOpenAPosition {
		if len(s.positions) == 0 {
			s.savePendingOrder("sell")
		}
	} else {
		if s.pendingOrder != nil {
			s.createPendingOrder("sell")
			return
		}
		s.pendingOrder = nil
	}

	ignoreLastNCandles := 14
	riskPercentage := float64(1)
	stopLossDistance := 12
	takeProfitDistance := 27
	candlesAmountWithLowerPriceToBeConsideredBottom := 14
	tpDistanceShortForBreakEvenSL := 5

	if len(s.positions) > 0 {
		s.checkIfSLShouldBeMovedToBreakEven(float64(tpDistanceShortForBreakEvenSL), "sell")
		return
	}

	lastCandlesIndex := len(s.candles) - 1
	for i := lastCandlesIndex - ignoreLastNCandles; i > lastCandlesIndex-ignoreLastNCandles-lastCandlesIndex; i-- {
		if i < 1 {
			break
		}
		isFalsePositive := false
		for j := i + 1; j < lastCandlesIndex-1; j++ {
			if s.candles[j].Low <= s.candles[i].Low {
				isFalsePositive = true
				break
			}
		}

		if isFalsePositive {
			s.Logger.Log("Doing nothing, not the proper trade setup for a short 1")
			break
		}

		isFalsePositive = false
		for j := i - candlesAmountWithLowerPriceToBeConsideredBottom; j < i; j++ {
			if j < 1 || j > lastCandlesIndex {
				continue
			}
			if s.candles[j].Low <= s.candles[i].Low {
				isFalsePositive = true
				break
			}
		}

		if isFalsePositive {
			s.Logger.Log("Doing nothing, not the proper trade setup for a short 2")
			break
		}

		price := s.candles[i].Low + 3
		if price >= float64(s.currentBrokerQuote.Bid) {
			s.Logger.Log("Price is lower than the current ask, so we can't create the short order now. Price is -> " + utils.FloatToString(price, 2))
			s.Logger.Log("Quote is -> " + utils.GetStringRepresentation(s.currentBrokerQuote))
			continue
		}

		// todo -> find a better name for closingOrdersTimestamp
		if s.closingOrdersTimestamp == s.candles[lastCandlesIndex].Timestamp {
			return
		}

		s.closingOrdersTimestamp = s.candles[lastCandlesIndex].Timestamp
		workingOrders := utils.GetWorkingOrders(s.orders)
		var ordersArr []*api.Order
		for _, order := range workingOrders {
			if order.Side == "sell" && order.ParentID == nil {
				ordersArr = append(ordersArr, order)
			}
		}
		s.Logger.Log("Ok, we might have a short setup, closing all working orders first if any ...")
		s.closeSpecificOrders(
			ordersArr,
			func() {
				highestValue := s.candles[lastCandlesIndex-1].High
				for i := lastCandlesIndex - 1; i > lastCandlesIndex-120; i-- {
					if i < 1 {
						break
					}
					if s.candles[i].High > highestValue {
						highestValue = s.candles[i].High
					}
				}
				diff := highestValue - s.candles[lastCandlesIndex-1].High
				if diff < 29 {
					s.Logger.Log("At the end it wasn't a good short setup, doing nothing ...")
					return
				}

				s.closeAllWorkingOrders(
					func() {
						float32Price := float32(price)

						stopLoss := float32Price + float32(stopLossDistance)
						takeProfit := float32Price - float32(takeProfitDistance)
						size := math.Floor((s.state.Equity*(riskPercentage/100))/float64(stopLossDistance) + 1)
						if size == 0 {
							size = 1
						}

						order := &api.Order{
							CurrentAsk: &s.currentBrokerQuote.Ask,
							CurrentBid: &s.currentBrokerQuote.Bid,
							Instrument: ibroker.GER30SymbolName,
							StopPrice:  &float32Price,
							Qty:        float32(size),
							Side:       "sell",
							StopLoss:   &stopLoss,
							TakeProfit: &takeProfit,
							Type:       "stop",
						}

						s.Logger.Log("Short order to be created -> " + utils.GetStringRepresentation(order))

						if !isValidTimeToOpenAPosition {
							s.Logger.Log("Now is not the time for opening any short orders, saving it for later ...")
							s.pendingOrder = order
						} else {
							s.createOrder(order, 20, 10*time.Second)
						}
					},
				)
			},
		)
	}
}

func (s *Strategy) isExecutionTimeValid(
	validMonths []string,
	validWeekDays []string,
	validHalfHours []string,
) bool {

	if len(validMonths) > 0 {
		if !utils.IsInArray(s.currentExecutionTime.Format("January"), validMonths) {
			return false
		}
	}

	if len(validWeekDays) > 0 {
		if !utils.IsInArray(s.currentExecutionTime.Format("Monday"), validWeekDays) {
			return false
		}
	}

	if len(validHalfHours) > 0 {
		currentHour, currentMinutes := s.getCurrentTimeHourAndMinutes()
		if currentMinutes >= 30 {
			currentMinutes = 30
		} else {
			currentMinutes = 0
		}

		currentHourString := strconv.Itoa(currentHour)
		currentMinutesString := strconv.Itoa(currentMinutes)
		if len(currentMinutesString) == 1 {
			currentMinutesString += "0"
		}

		return utils.IsInArray(currentHourString+":"+currentMinutesString, validHalfHours)
	}

	return true
}

func (s *Strategy) isCurrentTimeOutsideTradingHours() bool {
	currentHour, currentMinutes := s.getCurrentTimeHourAndMinutes()
	return (currentHour < 7) || (currentHour > 21) || (currentHour == 21 && currentMinutes > 57)
}

func (s *Strategy) getCurrentTimeHourAndMinutes() (int, int) {
	currentHour, _ := strconv.Atoi(s.currentExecutionTime.Format("15"))
	currentMinutes, _ := strconv.Atoi(s.currentExecutionTime.Format("04"))

	return currentHour, currentMinutes
}

func (s *Strategy) savePendingOrder(side string) {
	workingOrders := utils.GetWorkingOrders(s.orders)

	if len(workingOrders) == 0 {
		return
	}

	var mainOrder *api.Order
	for _, workingOrder := range workingOrders {
		if workingOrder.Side == side && workingOrder.ParentID == nil {
			mainOrder = workingOrder
		}
	}

	if mainOrder == nil {
		return
	}

	// TODO: savingPendingOrderTimestamp

	s.Logger.Log("Closing the current order and saving it for the future, since now it's not the time for profitable trading.")
	s.Logger.Log("This is the current order -> " + utils.GetStringRepresentation(mainOrder))

	slOrder, tpOrder := s.getSlAndTpOrders(mainOrder.ID, workingOrders)

	if slOrder != nil {
		mainOrder.StopLoss = slOrder.StopPrice
	}
	if tpOrder != nil {
		mainOrder.TakeProfit = tpOrder.LimitPrice
	}

	if mainOrder.Type == "limit" {
		mainOrder.StopPrice = nil
	}
	if mainOrder.Type == "stop" {
		mainOrder.LimitPrice = nil
	}

	s.closeAllWorkingOrders(func() {
		s.orders = nil
		s.pendingOrder = mainOrder
		s.Logger.Log("Closed all working orders correctly and pending order saved -> " + utils.GetStringRepresentation(s.pendingOrder))
	})
}

func (s *Strategy) createPendingOrder(side string) {
	s.Logger.Log("Trying to create the pending order ..." + utils.GetStringRepresentation(s.pendingOrder))
	var orderPrice float64
	if s.pendingOrder.Type == "limit" {
		orderPrice = float64(*s.pendingOrder.LimitPrice)
	} else {
		orderPrice = float64(*s.pendingOrder.StopPrice)
	}

	if (side == "buy" && s.pendingOrder.Type == "stop") || (side == "sell" && s.pendingOrder.Type == "limit") {
		if s.pendingOrder.Side == "buy" && orderPrice <= float64(s.currentBrokerQuote.Ask) {
			s.Logger.Log("Pending order will not be created since the order price is less than the current ask")
			s.pendingOrder = nil
			return
		}
		if side == "buy" && s.pendingOrder.Side == "sell" && s.pendingOrder.Type == "stop" {
			s.Logger.Log("Transforming sell stop order into a limit order -> " + utils.GetStringRepresentation(s.pendingOrder))
			s.pendingOrder.Type = "limit"
			s.pendingOrder.LimitPrice = s.pendingOrder.StopPrice
			s.pendingOrder.StopPrice = nil
		}
	}

	if (side == "buy" && s.pendingOrder.Type == "limit") || (side == "sell" && s.pendingOrder.Type == "stop") {
		if s.pendingOrder.Side == "sell" && orderPrice >= float64(s.currentBrokerQuote.Bid) {
			s.Logger.Log("Pending order will not be created since the order price is higher than the current bid")
			s.pendingOrder = nil
			return
		}
		if side == "sell" && s.pendingOrder.Side == "buy" && s.pendingOrder.Type == "stop" {
			s.Logger.Log("Transforming buy stop order into a limit order -> " + utils.GetStringRepresentation(s.pendingOrder))
			s.pendingOrder.Type = "limit"
			s.pendingOrder.LimitPrice = s.pendingOrder.StopPrice
			s.pendingOrder.StopPrice = nil
		}
	}

	// todo: creatependingordertimestamp

	s.createOrder(s.pendingOrder, 20, 10*time.Second)
	s.pendingOrder = nil
}

func (s *Strategy) getSlAndTpOrders(
	parentID string,
	orders []*api.Order,
) (*api.Order, *api.Order) {
	var slOrder *api.Order
	var tpOrder *api.Order
	for _, workingOrder := range orders {
		if workingOrder.ParentID == nil || *workingOrder.ParentID != parentID {
			continue
		}

		if workingOrder.Type == "limit" {
			tpOrder = workingOrder
		}
		if workingOrder.Type == "stop" {
			slOrder = workingOrder
		}
	}

	return slOrder, tpOrder
}

func (s *Strategy) getSlAndTpOrdersForCurrentOpenPosition() (
	slOrder *api.Order,
	tpOrder *api.Order,
) {
	for _, order := range s.orders {
		if order.Status != "working" {
			continue
		}
		if order.Type == "limit" {
			tpOrder = order
		}
		if order.Type == "stop" {
			slOrder = order
		}
	}
	return
}

func (s *Strategy) setStringValues(order *api.Order) {
	currentAsk := utils.FloatToString(float64(s.currentBrokerQuote.Ask), 1)
	currentBid := utils.FloatToString(float64(s.currentBrokerQuote.Bid), 1)
	qty := utils.IntToString(int64(order.Qty))
	order.StringValues = &api.OrderStringValues{
		CurrentAsk: &currentAsk,
		CurrentBid: &currentBid,
		Qty:        &qty,
	}

	if order.Type == "limit" {
		limitPrice := utils.FloatToString(math.Round(float64(*order.LimitPrice)*10)/10, 1)
		order.StringValues.LimitPrice = &limitPrice
	} else {
		stopPrice := utils.FloatToString(math.Round(float64(*order.StopPrice)*10)/10, 1)
		order.StringValues.StopPrice = &stopPrice
	}
	if order.StopLoss != nil {
		stopLossPrice := utils.FloatToString(math.Round(float64(*order.StopLoss)*10)/10, 1)
		order.StringValues.StopLoss = &stopLossPrice
	}
	if order.TakeProfit != nil {
		takeProfitPrice := utils.FloatToString(math.Round(float64(*order.TakeProfit)*10)/10, 1)
		order.StringValues.TakeProfit = &takeProfitPrice
	}
}

func (s *Strategy) checkIfSLShouldBeMovedToBreakEven(distanceToTp float64, side string) {
	if s.modifyingPositionTimestamp == s.getLastCandle().Timestamp {
		return
	}

	position := s.positions[0]
	if position.Side != side {
		return
	}

	_, tpOrder := s.getSlAndTpOrdersForCurrentOpenPosition()

	if tpOrder == nil {
		return
	}

	shouldBeAdjusted := false
	if side == "buy" {
		shouldBeAdjusted = float64(*tpOrder.LimitPrice)-s.getLastCandle().High < distanceToTp
	} else {
		shouldBeAdjusted = s.getLastCandle().Low-float64(*tpOrder.LimitPrice) < distanceToTp
	}

	if shouldBeAdjusted {
		s.Logger.Log("The trade is very close to the TP. Adjusting SL to break even ...")
		s.modifyingPositionTimestamp = s.getLastCandle().Timestamp

		s.modifyPosition(
			ibroker.GER30SymbolName,
			utils.FloatToString(float64(*tpOrder.LimitPrice), 2),
			utils.FloatToString(float64(position.AvgPrice), 2),
		)
	}
}
