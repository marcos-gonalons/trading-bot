package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/utils"
	"time"
)

func (s *Strategy) login(maxRetries int, delayBetweenRetries time.Duration) {
	go utils.RepeatUntilSuccess(
		"Login",
		func() (err error) {
			_, err = s.API.Login()
			if err != nil {
				s.Logger.Error("Error while logging in -> " + err.Error())
			}
			return
		},
		delayBetweenRetries,
		maxRetries,
		func() {},
	)
}

func (s *Strategy) closeSpecificOrders(
	orders []*api.Order,
	successCallback func(),
) {
	if orders == nil || len(orders) == 0 {
		successCallback()
		return
	}

	s.Logger.Log("Closing specified orders -> " + utils.GetStringRepresentation(orders))

	go utils.RepeatUntilSuccess(
		"CloseSpecifiedOrders",
		func() (err error) {
			for _, order := range orders {
				err = s.API.CloseOrder(order.ID)
				if err != nil {
					s.Logger.Error("An error happened while closing the specified orders -> " + err.Error())
					return
				}
			}
			return
		},
		5*time.Second,
		30,
		successCallback,
	)
}

func (s *Strategy) closeAllWorkingOrders(
	successCallback func(),
) {
	s.Logger.Log("Closing all working orders ...")

	go utils.RepeatUntilSuccess(
		"CloseAllWorkingOrders",
		func() (err error) {
			err = s.API.CloseAllOrders()
			if err != nil {
				s.Logger.Error("An error happened while closing all orders -> " + err.Error())
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

func (s *Strategy) modifyPosition(
	symbol string,
	tp string,
	sl string,
) {
	s.Logger.Log("Modifying the current open position with this values: symbol -> " + symbol + ", tp -> " + tp + " and sl -> " + sl)

	go utils.RepeatUntilSuccess(
		"ModifyPosition",
		func() (err error) {
			err = s.API.ModifyPosition(symbol, &tp, &sl)
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

func (s *Strategy) createOrder(
	order *api.Order,
	maxRetries int,
	delayBetweenRetries time.Duration,
) {
	s.Logger.Log("Creating this order -> " + utils.GetStringRepresentation(order))

	go utils.RepeatUntilSuccess(
		"CreateOrder",
		func() (err error) {
			if order.Side == "buy" {
				if order.Type == "limit" && *order.LimitPrice >= s.currentBrokerQuote.Bid {
					s.Logger.Log("Can't create the limit buy order since the order price is bigger than the current bid")
					return
				}
				if order.Type == "stop" && *order.StopPrice <= s.currentBrokerQuote.Ask {
					s.Logger.Log("Can't create the stop buy order since the order price is lower than the current ask")
					return
				}
			}
			if order.Side == "sell" {
				if order.Type == "limit" && *order.LimitPrice <= s.currentBrokerQuote.Ask {
					s.Logger.Log("Can't create the limit sell order since the order price is lower than the current ask")
					return
				}
				if order.Type == "stop" && *order.StopPrice >= s.currentBrokerQuote.Bid {
					s.Logger.Log("Can't create the stop sell order since the order price is bigger than the current bid")
					return
				}
			}

			s.setStringValues(order)
			err = s.API.CreateOrder(order)
			if err != nil {
				s.Logger.Error("Error when creating the order -> " + err.Error())
			} else {
				s.Logger.Log("Order created successfully")
			}
			return
		},
		10*time.Second,
		20,
		func() {},
	)
}
