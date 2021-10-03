package ibroker

import (
	"TradingBot/src/services/api"
	"strings"
)

// IsSessionDisconnectedError ...
func (s *API) IsSessionDisconnectedError(err error) bool {
	if err == nil {
		return false
	}

	for _, str := range SessionDisconnectedErrorStrings {
		if strings.Contains(err.Error(), str) {
			return true
		}
	}

	return false
}

// IsOrderAlreadyExistsError ...
func (s *API) IsOrderAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), OrderAlreadyExistsErrorString)
}

// IsNotEnoughFundsError ...
func (s *API) IsNotEnoughFundsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), NotEnoughFundsErrorString)
}

// IsOrderPendingCancelError ...
func (s *API) IsOrderPendingCancelError(err error) bool {
	return err != nil && strings.Contains(err.Error(), OrderIsPendingCancelErrorString)
}

// IsOrderCancelledError ...
func (s *API) IsOrderCancelledError(err error) bool {
	return err != nil && strings.Contains(err.Error(), OrderIsCancelledErrorString)
}

// IsOrderFilledError ...
func (s *API) IsOrderFilledError(err error) bool {
	return err != nil && strings.Contains(err.Error(), OrderIsFilledErrorString)
}

// IsInvalidHoursError ...
func (s *API) IsInvalidHoursError(err error) bool {
	return err != nil && strings.Contains(err.Error(), InvalidHoursErrorString)
}

// IsClosePositionRequestInProgressError ...
func (s *API) IsClosePositionRequestInProgressError(err error) bool {
	return err != nil && strings.Contains(err.Error(), ClosePositionRequestInProgressErrorString)
}

// IsLimitOrder ...
func (s *API) IsLimitOrder(order *api.Order) bool {
	return order.Type == LimitType
}

// IsStopOrder ...
func (s *API) IsStopOrder(order *api.Order) bool {
	return order.Type == StopType
}

// IsLongOrder ...
func (s *API) IsLongOrder(order *api.Order) bool {
	return order.Side == LongSide
}

// IsShortOrder ...
func (s *API) IsShortOrder(order *api.Order) bool {
	return order.Side == ShortSide
}

// IsLongPosition ...
func (s *API) IsLongPosition(position *api.Position) bool {
	return position.Side == LongSide
}

// IsShortPosition ...
func (s *API) IsShortPosition(position *api.Position) bool {
	return position.Side == ShortSide
}

// IsWorkingOrder ...
func (s *API) IsWorkingOrder(order *api.Order) bool {
	return order.Status == StatusWorkingOrder
}
