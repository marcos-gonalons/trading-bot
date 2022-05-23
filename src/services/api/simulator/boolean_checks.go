package simulator

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/simulator/constants"
)

// IsSessionDisconnectedError ...
func (s *API) IsSessionDisconnectedError(err error) bool {
	return false
}

// IsOrderAlreadyExistsError ...
func (s *API) IsOrderAlreadyExistsError(err error) bool {
	return false
}

// IsPositionAlreadyExistsError ...
func (s *API) IsPositionAlreadyExistsError(err error) bool {
	return false
}

// IsNotEnoughFundsError ...
func (s *API) IsNotEnoughFundsError(err error) bool {
	return false
}

// IsOrderPendingCancelError ...
func (s *API) IsOrderPendingCancelError(err error) bool {
	return false
}

// IsOrderCancelledError ...
func (s *API) IsOrderCancelledError(err error) bool {
	return false
}

// IsOrderFilledError ...
func (s *API) IsOrderFilledError(err error) bool {
	return false
}

// IsInvalidHoursError ...
func (s *API) IsInvalidHoursError(err error) bool {
	return false
}

// IsClosePositionRequestInProgressError ...
func (s *API) IsClosePositionRequestInProgressError(err error) bool {
	return false
}

// IsPositionNotFoundError ...
func (s *API) IsPositionNotFoundError(err error) bool {
	return false
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
	return true
}
