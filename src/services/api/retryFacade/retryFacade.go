package retryFacade

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
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

	s.Logger.Log("Closing specified orders -> " + utils.GetStringRepresentation(orders))

	if len(orders) == 0 {
		retryParams.SuccessCallback()
		return
	}

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
		"ClosePositions",
		func() (err error) {
			err = s.API.CloseAllPositions()
			if err != nil {
				s.Logger.Error("An error happened while closing all positions -> " + err.Error())
			}
			if s.API.IsInvalidHoursError(err) || s.API.IsClosePositionRequestInProgressError(err) {
				err = nil
			}
			return
		},
		retryParams.DelayBetweenRetries,
		retryParams.MaxRetries,
		retryParams.SuccessCallback,
	)
}

// ClosePosition ...
func (s *APIFacade) ClosePosition(symbol string, retryParams RetryParams) {
	s.Logger.Log("Closing the specified position ..." + symbol)

	go utils.RepeatUntilSuccess(
		"ClosePosition",
		func() (err error) {
			err = s.API.ClosePosition(symbol)
			if err != nil {
				s.Logger.Error("An error happened while closing the position for " + symbol + " -> " + err.Error())
			}
			if s.API.IsInvalidHoursError(err) {
				err = nil
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
				if s.API.IsPositionNotFoundError(err) || s.API.IsInvalidHoursError(err) {
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

			if s.API.IsLongOrder(order) {
				if s.API.IsLimitOrder(order) && *order.LimitPrice >= currentQuote.Bid {
					s.Logger.Log("Can't create the limit buy order since the order price is bigger than the current bid")
					return
				}
				if s.API.IsStopOrder(order) && *order.StopPrice <= currentQuote.Ask {
					s.Logger.Log("Can't create the stop buy order since the order price is lower than the current ask")
					return
				}
			}
			if s.API.IsShortOrder(order) {
				if s.API.IsLimitOrder(order) && *order.LimitPrice <= currentQuote.Ask {
					s.Logger.Log("Can't create the limit sell order since the order price is lower than the current ask")
					return
				}
				if s.API.IsStopOrder(order) && *order.StopPrice >= currentQuote.Bid {
					s.Logger.Log("Can't create the stop sell order since the order price is bigger than the current bid")
					return
				}
			}

			setStringValues(order)
			err = s.API.CreateOrder(order)
			if err != nil {
				s.Logger.Error("Error when creating the order -> " + err.Error())
				// todo: group this errors into 'known errors' or 'acceptable errors'
				if s.API.IsOrderAlreadyExistsError(err) || s.API.IsNotEnoughFundsError(err) || s.API.IsPositionAlreadyExistsError(err) || s.API.IsInvalidHoursError(err) {
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
