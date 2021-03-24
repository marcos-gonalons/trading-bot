package retryFacade

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
	"strings"
)

// APIFacade ...
type APIFacade struct {
	API    api.Interface
	Logger logger.Interface
}

// Login ...
func (s *APIFacade) Login(retryParams RetryParams) {
	go utils.RepeatUntilSuccess(
		"Login",
		func() (err error) {
			_, err = s.API.Login()
			if err != nil {
				s.Logger.Error("Error while logging in -> " + err.Error())
			}
			return
		},
		retryParams.DelayBetweenRetries,
		retryParams.MaxRetries,
		func() {},
	)
}

// CloseOrders ...
func (s *APIFacade) CloseOrders(
	orders []*api.Order,
	retryParams RetryParams,
) {

	if len(orders) == 0 {
		retryParams.SuccessCallback()
		return
	}

	s.Logger.Log("Closing specified orders -> " + utils.GetStringRepresentation(orders))

	go utils.RepeatUntilSuccess(
		"CloseOrders",
		func() (err error) {

			for _, order := range orders {
				err = s.API.CloseOrder(order.ID)

				if err != nil {
					s.Logger.Error("An error happened while closing this order -> " + utils.GetStringRepresentation(order))
					s.Logger.Error("Error -> " + err.Error())

					if s.API.IsOrderPendingCancelError(err) || s.API.IsOrderCancelledError(err) || s.API.IsOrderFilledError(err) {
						err = nil
					} else {
						return
					}
				}
			}
			return
		},
		retryParams.DelayBetweenRetries,
		retryParams.MaxRetries,
		retryParams.SuccessCallback,
	)
}

// ClosePositions ...
func (s *APIFacade) ClosePositions(retryParams RetryParams) {
	s.Logger.Log("Closing all positions ...")

	go utils.RepeatUntilSuccess(
		"CloseAllPositions",
		func() (err error) {
			err = s.API.CloseAllPositions()
			if err != nil {
				s.Logger.Error("An error happened while closing all positions -> " + err.Error())
				retryParams.ErrorCallback(err)
			}
			return
		},
		retryParams.DelayBetweenRetries,
		retryParams.MaxRetries,
		retryParams.SuccessCallback,
	)
}

// ModifyPosition ...
func (s *APIFacade) ModifyPosition(
	symbol string,
	tp string,
	sl string,
	retryParams RetryParams,
) {
	s.Logger.Log("Modifying the current open position with this values: symbol -> " + symbol + ", tp -> " + tp + " and sl -> " + sl)

	go utils.RepeatUntilSuccess(
		"ModifyPosition",
		func() (err error) {
			err = s.API.ModifyPosition(symbol, &tp, &sl)
			if err != nil {
				s.Logger.Error("An error happened while modifying the position -> " + err.Error())
				if strings.Contains(err.Error(), "no tiene posición abierta en el contrato") {
					err = nil
				}
			}
			return
		},
		retryParams.DelayBetweenRetries,
		retryParams.MaxRetries,
		retryParams.SuccessCallback,
	)
}

// CreateOrder ...
func (s *APIFacade) CreateOrder(
	order *api.Order,
	getCurrentBrokerQuote func() *api.Quote,
	setStringValues func(order *api.Order),
	retryParams RetryParams,
) {
	s.Logger.Log("Creating this order -> " + utils.GetStringRepresentation(order))

	go utils.RepeatUntilSuccess(
		"CreateOrder",
		func() (err error) {
			currentQuote := getCurrentBrokerQuote()
			s.Logger.Log("Current broker quote is" + utils.GetStringRepresentation(currentQuote))

			if order.Side == "buy" {
				if order.Type == "limit" && *order.LimitPrice >= currentQuote.Bid {
					s.Logger.Log("Can't create the limit buy order since the order price is bigger than the current bid")
					return
				}
				if order.Type == "stop" && *order.StopPrice <= currentQuote.Ask {
					s.Logger.Log("Can't create the stop buy order since the order price is lower than the current ask")
					return
				}
			}
			if order.Side == "sell" {
				if order.Type == "limit" && *order.LimitPrice <= currentQuote.Ask {
					s.Logger.Log("Can't create the limit sell order since the order price is lower than the current ask")
					return
				}
				if order.Type == "stop" && *order.StopPrice >= currentQuote.Bid {
					s.Logger.Log("Can't create the stop sell order since the order price is bigger than the current bid")
					return
				}
			}

			setStringValues(order)
			err = s.API.CreateOrder(order)
			if err != nil {
				s.Logger.Error("Error when creating the order -> " + err.Error())
				if s.API.IsOrderAlreadyExistsError(err) || s.API.IsNotEnoughFundsError(err) {
					err = nil
				}
			} else {
				s.Logger.Log("Order created successfully")
			}
			return
		},
		retryParams.DelayBetweenRetries,
		retryParams.MaxRetries,
		retryParams.SuccessCallback,
	)
}

// CreateAPIFacadeInstance ...
func CreateAPIFacadeInstance(api api.Interface) Interface {
	instance := &APIFacade{
		API:    api,
		Logger: logger.GetInstance(),
	}

	return instance
}
