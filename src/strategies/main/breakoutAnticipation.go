package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/utils"
	"strconv"
	"time"
)

func (s *Strategy) breakoutAnticipationStrategy() {
	if s.isCurrentTimeOutsideTradingHours() {
		s.Logger.Log("Doing nothing - Now it's not the time.")
		s.closeWorkingOrders(
			func() {
				s.orders = nil
				s.closePositions(func() { s.positions = nil }, func(err error) {})
			},
			func(err error) {},
		)

		return
	}

	if s.averageSpread > 3 {
		s.Logger.Log("Doing nothing since the spread is very big -> " + utils.FloatToString(s.averageSpread, 0))
		s.pendingOrder = nil
		s.closeWorkingOrders(func() { s.orders = nil }, func(err error) {})
		return
	}

	//s.resistanceBreakoutAnticipationStrategy()
	//s.supportBreakoutAnticipationStrategy()
}

func (s *Strategy) resistanceBreakoutAnticipationStrategy() {
	isValidTime := s.isExecutionTimeValid(
		[]string{"January", "March", "April", "May", "June", "August", "September", "October"},
		[]string{"Monday", "Tuesday", "Wednesday", "Friday"},
		[]string{"9:00", "10:00", "11:30", "12:00", "12:30", "20:30"},
	)

	if !isValidTime {
		s.savePendingOrder("buy")
	} else {
		if s.pendingOrder != nil {
			s.createPendingOrder("buy")
			return
		}
		s.pendingOrder = nil
	}

	/*ignoreLastNCandles := 15
	riskPercentage := 1.5
	stopLossDistance := 12
	takeProfitDistance := 27
	candlesAmountWithLowerPriceToBeConsideredTop := 15*/
	tpDistanceShortForBreakEvenSL := 5

	if len(s.positions) > 0 {
		s.checkIfSLShouldBeMovedToBreakEven(float64(tpDistanceShortForBreakEvenSL), "buy")
	}

}

func (s *Strategy) supportBreakoutAnticipationStrategy() {
	isValidTime := s.isExecutionTimeValid(
		[]string{"March", "April", "June", "September", "December"},
		[]string{"Monday", "Tuesday", "Thursday", "Friday"},
		[]string{"08:30", "9:00", "12:00", "13:00", "14:30", "15:30", "18:00"},
	)

	if !isValidTime {
		s.savePendingOrder("sell")
	} else {
		if s.pendingOrder != nil {
			s.createPendingOrder("sell")
			return
		}
		s.pendingOrder = nil
	}

	/*ignoreLastNCandles := 15
	riskPercentage := 1.5
	stopLossDistance := 12
	takeProfitDistance := 27
	candlesAmountWithLowerPriceToBeConsideredBottom := 15*/
	tpDistanceShortForBreakEvenSL := 5

	if len(s.positions) > 0 {
		s.checkIfSLShouldBeMovedToBreakEven(float64(tpDistanceShortForBreakEvenSL), "sell")
	}
}

func (s *Strategy) isExecutionTimeValid(
	validMonths []string,
	validWeekDays []string,
	validHalfHours []string,
) bool {

	if !isInArray(s.currentExecutionTime.Format("January"), validMonths) {
		return false
	}

	if !isInArray(s.currentExecutionTime.Format("Monday"), validWeekDays) {
		return false
	}

	currentHour, currentMinutes := s.getCurrentTimeHourAndMinutes()
	if currentMinutes > 30 {
		currentMinutes = 30
	}

	return isInArray(strconv.Itoa(currentHour)+":"+strconv.Itoa(currentMinutes), validHalfHours)
}

func (s *Strategy) isCurrentTimeOutsideTradingHours() bool {
	currentHour, currentMinutes := s.getCurrentTimeHourAndMinutes()
	return (currentHour < 6) || (currentHour > 21) || (currentHour == 21 && currentMinutes > 57)
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
	if mainOrder != nil {
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
		s.Logger.Log("Pending order saved -> " + utils.GetStringRepresentation(s.pendingOrder))

		s.closeWorkingOrders(func() {
			s.orders = nil
			s.pendingOrder = mainOrder
			s.Logger.Log("Closed all orders correctly, and saved the previous order for later")
			s.Logger.Log("Closing the position now ...")
			s.closePositions(func() { s.positions = nil }, func(err error) {})
		}, func(err error) { s.pendingOrder = nil })

	}
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

	s.setStringValues(s.pendingOrder)

	err := s.API.CreateOrder(s.pendingOrder)
	if err != nil {
		s.Logger.Error("Error when creating the pending order -> " + err.Error())
	} else {
		s.Logger.Log("Pending order created successfully")
	}

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

func (s *Strategy) closeWorkingOrders(
	successCallback func(),
	onErrorCallback func(err error),
) {
	s.Logger.Log("Closing all working orders ...")

	go utils.RepeatUntilSuccess(
		"CloseAllWorkingOrders",
		func() (err error) {
			err = s.API.CloseAllOrders()
			if err != nil {
				s.Logger.Error("An error happened while closing all orders -> " + err.Error())
				onErrorCallback(err)
			}
			return
		},
		5*time.Second,
		30,
		successCallback,
	)
}

func (s *Strategy) closePositions(
	successCallback func(),
	onErrorCallback func(err error),
) {
	s.Logger.Log("Closing all positions ...")

	go utils.RepeatUntilSuccess(
		"CloseAllPositions",
		func() (err error) {
			err = s.API.CloseAllPositions()
			if err != nil {
				s.Logger.Error("An error happened while closing all positions -> " + err.Error())
				onErrorCallback(err)
			}
			return
		},
		5*time.Second,
		30,
		successCallback,
	)
}

func (s *Strategy) modifyPosition(tp string, sl string) {
	s.Logger.Log("Modifying the current open position with this values: tp -> " + tp + " and sl -> " + sl)

	go utils.RepeatUntilSuccess(
		"ModifyPosition",
		func() (err error) {
			err = s.API.ModifyPosition(ibroker.GER30SymbolName, &tp, &sl)
			if err != nil {
				s.Logger.Error("An error happened while modifying the position -> " + err.Error())
			}
			return
		},
		5*time.Second,
		20,
		func() {},
	)
}

func (s *Strategy) setStringValues(order *api.Order) {
	currentAsk := utils.FloatToString(float64(s.currentBrokerQuote.Ask), 2)
	currentBid := utils.FloatToString(float64(s.currentBrokerQuote.Bid), 2)
	qty := utils.FloatToString(float64(order.Qty), 2)
	order.StringValues = &api.OrderStringValues{
		CurrentAsk: &currentAsk,
		CurrentBid: &currentBid,
		Qty:        &qty,
	}
	if order.Type == "limit" {
		limitPrice := utils.FloatToString(float64(*order.LimitPrice), 2)
		order.StringValues.LimitPrice = &limitPrice
	} else {
		stopPrice := utils.FloatToString(float64(*order.StopPrice), 2)
		order.StringValues.StopPrice = &stopPrice
	}
	if order.StopLoss != nil {
		stopLossPrice := utils.FloatToString(float64(*order.StopLoss), 2)
		order.StringValues.StopLoss = &stopLossPrice
	}
	if order.TakeProfit != nil {
		takeProfitPrice := utils.FloatToString(float64(*order.TakeProfit), 2)
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
			utils.FloatToString(float64(*tpOrder.LimitPrice), 2),
			utils.FloatToString(float64(position.AvgPrice), 2),
		)
	}
}

func isInArray(element string, arr []string) bool {
	for _, el := range arr {
		if element == el {
			return true
		}
	}
	return false
}

/**

	// When it's time to create the order:
		create order code {

			if creatingOrderTimestamp != currentCandle.timestamp {
				createOrder
				s.creatingOrderTimestamp = now (with 00 seconds)
			}

		}

	// when it's time to modify the sl/tp of an order
		modifyingOrderTimestamp

	// When it's time to modify the sl/tp of a position


	fx va 1 pip por detras, mas o menos.
	si fx:ger30 dice 13149.4, en ibroker es 13150.5

	Hoy domingo, mercado cerrado -> FX:GER30 dixe 13123.4, IBROKER:GER30 dice 13123.8
	Sin embargo, los picos siempre muestran 1 pip de diferencia
Por ahora, asumir que la diferencia entre uno y otro es de 1 pip

	leer con unauthorized de fx:ger30
	y tener en cuenta que va 1 por detras


	Important; the script should ignore candles[0], since it does not contain proper data.

	Very, very important
	Round ALL the prices used in ALL the calls to 2 decimals. Otherwise it won't work.

	When creating an order, I need to save the 3 created orders somewhere (the limit/stop order, it's sl and it's tp)
	The SL and the TP will have the parentID of the main one. The main one will have the parentID null
	All 3 orders will have the status "working".

	When modifying an order that hasn't been filled yet, I can use the ID of the main order to change it's sl, tp, or it's limit/stop price.
	When modifying the sl/tp of a position, I need to use the ID of the sl/tp order.
	Or I can just use the modifyposition api


	Take into consideration
	Let's say the bot dies, for whatever reason, at 15:00pm
	I revive him at 15:05
	It will have lost all the candles[]

	To mitigate this
	As I add a candle to the candles[]
	Save the candles to the csv file
	When booting the bot; initialize the candles array with those in the csv file


	When booting {
		if !csv file, create the csv file
		else, load candles[] from the file
	}

	Very important as well
	When closing a position, and the position has TP and SL; IBROKER WILL NOT LET YOU
	CLOSE THE POSITION UNTIL YOU CLOSE THE TP AND SL FIRST
	So; if the script needs to close a position, CLOSE THE SL AND TP FIRST.


	Also very important; when the order is created,
	I need to quickly check it's SL and TP, and adjust them correctly to 12/27
***/
