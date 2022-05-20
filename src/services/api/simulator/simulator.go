package simulator

import (
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
	"errors"
	"time"

	"TradingBot/src/services/api"
)

// API ...
type API struct {
	logger logger.Interface

	state     *api.State
	orders    []*api.Order
	positions []*api.Position
}

// Login ...
func (s *API) Login() (accessToken *api.AccessToken, err error) {
	return nil, nil
}

// GetQuote ...
func (s *API) GetQuote(symbol string) (quote *api.Quote, err error) {
	return nil, nil
}

// CreateOrder ...
func (s *API) CreateOrder(order *api.Order) (err error) {
	for _, o := range s.orders {
		if o.Instrument == order.Instrument {
			if s.IsLimitOrder(o) {
				err = errors.New("there is already a limit order for this instrument")
				return
			}
			if s.IsStopOrder(o) {
				err = errors.New("there is already a stop order for this instrument")
				return
			}
		}
	}

	order.ID = utils.GetRandomString(5)
	s.orders = append(s.orders, order)

	bracketOrdersSide := ""
	if order.Side == "buy" {
		bracketOrdersSide = "sell"
	} else if order.Side == "sell" {
		bracketOrdersSide = "buy"
	}
	if order.TakeProfit != nil {
		takeProfit := api.Order{}
		takeProfit.Qty = order.Qty
		takeProfit.ID = utils.GetRandomString(5)
		takeProfit.Type = "limit"
		takeProfit.LimitPrice = order.TakeProfit
		takeProfit.ParentID = &order.ID
		takeProfit.Instrument = order.Instrument
		takeProfit.Side = bracketOrdersSide
		s.orders = append(s.orders, &takeProfit)
	}

	if order.StopLoss != nil {
		stopLoss := api.Order{}
		stopLoss.Qty = order.Qty
		stopLoss.ID = utils.GetRandomString(5)
		stopLoss.Type = "stop"
		stopLoss.StopPrice = order.StopPrice
		stopLoss.ParentID = &order.ID
		stopLoss.Instrument = order.Instrument
		stopLoss.Side = bracketOrdersSide
		s.orders = append(s.orders, &stopLoss)
	}
	return nil
}

// GetOrders ...
func (s *API) GetOrders() (orders []*api.Order, err error) {
	return s.orders, nil
}

// ModifyOrder ...
func (s *API) ModifyOrder(order *api.Order) (err error) {
	// TODO: This method is not used anywhere, delete?
	return nil
}

// ClosePosition ...
func (s *API) ClosePosition(symbol string) (err error) {
	positionIndex := 0
	for index, position := range s.positions {
		if position.Instrument == symbol {
			positionIndex = index
			break
		}
	}
	s.positions = append(s.positions[:positionIndex], s.positions[positionIndex+1:]...)

	takeProfitOrderIndex := 0
	hasTakeProfit := false
	stopLossOrderIndex := 0
	hasStopLoss := false
	for index, order := range s.orders {
		if order.Instrument != symbol {
			continue
		}
		if order.LimitPrice != nil {
			takeProfitOrderIndex = index
			hasTakeProfit = true
		}
		if order.StopPrice != nil {
			stopLossOrderIndex = index
			hasStopLoss = true
		}
	}

	if hasTakeProfit {
		s.orders = append(s.orders[:takeProfitOrderIndex], s.orders[takeProfitOrderIndex+1:]...)
	}
	if hasStopLoss {
		s.orders = append(s.orders[:stopLossOrderIndex], s.orders[stopLossOrderIndex+1:]...)
	}
	return nil
}

// CloseOrder ...
func (s *API) CloseOrder(orderID string) (err error) {
	orderIndex := 0
	for index, order := range s.orders {
		if order.ID == orderID {
			orderIndex = index
			break
		}
	}

	s.orders = append(s.orders[:orderIndex], s.orders[orderIndex+1:]...)
	return nil
}

// GetPositions ...
func (s *API) GetPositions() (positions []*api.Position, err error) {
	return s.positions, nil
}

// GetState ...
func (s *API) GetState() (state *api.State, err error) {
	// todo: handle state
	return &api.State{
		Balance:      1000,
		UnrealizedPL: 0,
		Equity:       1000,
	}, nil
}

// ModifyPosition ...
func (s *API) ModifyPosition(symbol string, takeProfit *string, stopLoss *string) (err error) {
	var position *api.Position
	for _, p := range s.positions {
		if position.Instrument != symbol {
			continue
		}
		position = p
	}
	if position == nil {
		return errors.New("position not found")
	}

	hasTP := false
	hasSL := false
	for _, o := range s.orders {
		if o.Instrument != symbol {
			continue
		}

		if s.IsLimitOrder(o) {
			var aux float32
			var v *float32
			hasTP = true
			if *takeProfit != "" {
				aux = float32(utils.StringToFloat(*takeProfit))
			}
			v = &aux
			o.LimitPrice = v
		}
		if s.IsStopOrder(o) {
			var aux float32
			var v *float32
			hasSL = true
			if *stopLoss != "" {
				aux = float32(utils.StringToFloat(*stopLoss))
			}
			v = &aux
			o.StopPrice = v
		}
	}

	var side string
	if position.Side == "buy" {
		side = "sell"
	} else {
		side = "buy"
	}
	if !hasTP {
		tpOrder := api.Order{}
		tpOrder.Qty = position.Qty
		tpOrder.ID = utils.GetRandomString(5)
		tpOrder.Type = "limit"
		tp := float32(utils.StringToFloat(*takeProfit))
		tpOrder.LimitPrice = &tp
		tpOrder.Instrument = position.Instrument
		tpOrder.Side = side
		s.CreateOrder(&tpOrder)
	}

	if !hasSL {
		slOrder := api.Order{}
		slOrder.Qty = position.Qty
		slOrder.ID = utils.GetRandomString(5)
		slOrder.Type = "stop"
		tp := float32(utils.StringToFloat(*stopLoss))
		slOrder.StopPrice = &tp
		slOrder.Instrument = position.Instrument
		slOrder.Side = side
		s.CreateOrder(&slOrder)
	}

	return nil
}

// GetWorkingOrders ...
func (s *API) GetWorkingOrders(orders []*api.Order) []*api.Order {
	return s.orders
}

// CloseAllOrders ...
func (s *API) CloseAllOrders() (err error) {
	s.orders = nil
	return nil
}

// CloseAllPositions ...
func (s *API) CloseAllPositions() (err error) {
	s.positions = nil
	return nil
}

// SetTimeout ...
func (s *API) SetTimeout(t time.Duration) {
}

// GetTimeout ...
func (s *API) GetTimeout() time.Duration {
	return 1
}

func (s *API) GetBracketOrdersForOpenedPosition(position *api.Position) (
	slOrder *api.Order,
	tpOrder *api.Order,
) {
	for _, order := range s.orders {
		if order.Instrument != position.Instrument {
			continue
		}
		if s.IsLimitOrder(order) {
			tpOrder = order
		}
		if s.IsStopOrder(order) {
			slOrder = order
		}
	}
	return
}

func (s *API) GetWorkingOrderWithBracketOrders(side string, symbol string, orders []*api.Order) []*api.Order {
	var workingOrders []*api.Order

	for _, order := range s.orders {
		if order.Side != side || order.Instrument != symbol || order.ParentID != nil {
			continue
		}

		workingOrders = append(workingOrders, order)
	}

	if len(workingOrders) > 0 {
		for _, order := range s.orders {
			if order.ParentID == nil || *order.ParentID != workingOrders[0].ID {
				continue
			}

			workingOrders = append(workingOrders, order)
		}
	}

	return workingOrders
}

// CreateAPIServiceInstance ...
func CreateAPIServiceInstance() api.Interface {
	instance := &API{
		logger: logger.GetInstance(),
	}

	return instance
}
