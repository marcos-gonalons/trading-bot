package ibroker

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker/constants"
	"strings"
)

// IsSessionDisconnectedError ...
func (s *API) IsSessionDisconnectedError(err error) bool {
	if err == nil {
		return false
	}

	for _, str := range constants.SessionDisconnectedErrorStrings {
		if strings.Contains(err.Error(), str) {
			return true
		}
	}

	return false
}

// IsOrderAlreadyExistsError ...
func (s *API) IsOrderAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), constants.OrderAlreadyExistsErrorString)
}

// IsPositionAlreadyExistsError ...
func (s *API) IsPositionAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), constants.PositionAlreadyExistsErrorString)
}

// IsNotEnoughFundsError ...
func (s *API) IsNotEnoughFundsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), constants.NotEnoughFundsErrorString)
}

// IsOrderPendingCancelError ...
func (s *API) IsOrderPendingCancelError(err error) bool {
	return err != nil && strings.Contains(err.Error(), constants.OrderIsPendingCancelErrorString)
}

// IsOrderCancelledError ...
func (s *API) IsOrderCancelledError(err error) bool {
	return err != nil && strings.Contains(err.Error(), constants.OrderIsCancelledErrorString)
}

// IsOrderFilledError ...
func (s *API) IsOrderFilledError(err error) bool {
	return err != nil && strings.Contains(err.Error(), constants.OrderIsFilledErrorString)
}

// IsInvalidHoursError ...
func (s *API) IsInvalidHoursError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), constants.InvalidHoursErrorString) || strings.Contains(err.Error(), constants.InvalidHoursErrorString2))
}

// IsClosePositionRequestInProgressError ...
func (s *API) IsClosePositionRequestInProgressError(err error) bool {
	return err != nil && strings.Contains(err.Error(), constants.ClosePositionRequestInProgressErrorString)
}

// IsLimitOrder ...
func (s *API) IsLimitOrder(order *api.Order) bool {
	return order.Type == constants.LimitType
}

// IsStopOrder ...
func (s *API) IsStopOrder(order *api.Order) bool {
	return order.Type == constants.StopType
}

// IsLongOrder ...
func (s *API) IsLongOrder(order *api.Order) bool {
	return order.Side == constants.LongSide
}

// IsShortOrder ...
func (s *API) IsShortOrder(order *api.Order) bool {
	return order.Side == constants.ShortSide
}

// IsLongPosition ...
func (s *API) IsLongPosition(position *api.Position) bool {
	return position.Side == constants.LongSide
}

// IsShortPosition ...
func (s *API) IsShortPosition(position *api.Position) bool {
	return position.Side == constants.ShortSide
}

// IsWorkingOrder ...
func (s *API) IsWorkingOrder(order *api.Order) bool {
	return order.Status == constants.StatusWorkingOrder
}
