package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/utils"
	"strconv"
)

func (s *Strategy) breakoutAnticipationStrategy() {
	var err error

	if s.isCurrentTimeOutsideTradingHours() {
		s.Logger.Log("Doing nothing - Now it's not the time.")
		if len(s.positions) > 0 || len(s.orders) > 0 {
			s.Logger.Log("Closing all open positions and pending orders...")
			err = s.API.CloseEverything()
			if err != nil {
				s.Logger.Log("An error happened while closing everything -> " + err.Error())
			} else {
				s.positions = nil
				s.orders = nil
			}
		}
		return
	}

	s.resistanceBreakoutAnticipationStrategy()
	s.supportBreakoutAnticipationStrategy()
}

func (s *Strategy) resistanceBreakoutAnticipationStrategy() {
	isValidTime := s.isExecutionTimeValid(
		[]string{"January", "March", "April", "May", "June", "August", "September", "October", "November", "December"},
		[]string{"Monday", "Tuesday", "Wednesday", "Friday"},
		[]string{"9:00", "9:30", "10:00", "11:30", "12:00", "12:30", "20:30"},
	)

	if !isValidTime {
		workingOrders := utils.GetWorkingOrders(s.orders)
		if len(workingOrders) > 0 {
			var mainOrder *api.Order
			for _, workingOrder := range workingOrders {
				if workingOrder.Side == "buy" && workingOrder.ParentID == nil {
					mainOrder = workingOrder
				}
			}
			if mainOrder != nil {
				s.API.CloseEverything()

				var slOrder *api.Order
				var tpOrder *api.Order
				for _, workingOrder := range workingOrders {
					if workingOrder.ParentID != nil && *workingOrder.ParentID == mainOrder.ID {
						if workingOrder.Type == "limit" {
							tpOrder = workingOrder
						}
						if workingOrder.Type == "stop" {
							slOrder = workingOrder
						}
					}
				}
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
				s.pendingOrder = mainOrder
			}
		}
	} else {
		if s.pendingOrder != nil {
			var price float32
			if s.pendingOrder.Type == "limit" {
				price = *s.pendingOrder.LimitPrice
			} else {
				price = *s.pendingOrder.StopPrice
			}
			if float64(price) > float64(s.currentBrokerQuote.Ask) {
				currentAsk := utils.FloatToString(float64(s.currentBrokerQuote.Ask), 2)
				currentBid := utils.FloatToString(float64(s.currentBrokerQuote.Bid), 2)
				qty := utils.FloatToString(float64(s.pendingOrder.Qty), 2)
				s.pendingOrder.StringValues = &api.OrderStringValues{
					CurrentAsk: &currentAsk,
					CurrentBid: &currentBid,
					Qty:        &qty,
				}
				if s.pendingOrder.Type == "limit" {
					limitPrice := utils.FloatToString(float64(*s.pendingOrder.LimitPrice), 2)
					s.pendingOrder.StringValues.LimitPrice = &limitPrice
				} else {
					stopPrice := utils.FloatToString(float64(*s.pendingOrder.StopPrice), 2)
					s.pendingOrder.StringValues.StopPrice = &stopPrice
				}
				if s.pendingOrder.StopLoss != nil {
					stopLossPrice := utils.FloatToString(float64(*s.pendingOrder.StopLoss), 2)
					s.pendingOrder.StringValues.StopLoss = &stopLossPrice
				}
				if s.pendingOrder.TakeProfit != nil {
					takeProfitPrice := utils.FloatToString(float64(*s.pendingOrder.TakeProfit), 2)
					s.pendingOrder.StringValues.TakeProfit = &takeProfitPrice
				}
				err := s.API.CreateOrder(s.pendingOrder)
				if err != nil {
					s.Logger.Log("Error when creating the order -> " + err.Error())
				}
			}
			s.pendingOrder = nil
			return
		}
		s.pendingOrder = nil
		// Important; when creating the order, I need to add the current ask and current bid to the string values
		// But before, check if I can create the order without adding that
	}
}

func (s *Strategy) supportBreakoutAnticipationStrategy() {

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

func isInArray(element string, arr []string) bool {
	for _, el := range arr {
		if element == el {
			return true
		}
	}
	return false
}

/**

if average spread is more than 3, close all orders if any, (leave the positions open if any) and do nothing

	fx va 1 pip por detras
	si fx:ger30 dice 13149.4, en ibroker es 13150.5

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
***/
