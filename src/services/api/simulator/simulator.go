package simulator

import (
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"errors"
	"math"
	"time"

	"TradingBot/src/services/api"
	apiConstants "TradingBot/src/services/api/simulator/constants"
)

// API ...
type API struct {
	logger logger.Interface

	state     *api.State
	orders    []*api.Order
	positions []*api.Position

	trades []*api.Trade
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
	if s.IsLongOrder(order) {
		bracketOrdersSide = apiConstants.ShortSide
	} else if s.IsShortOrder(order) {
		bracketOrdersSide = apiConstants.LongSide
	}
	if order.TakeProfit != nil {
		takeProfit := api.Order{}
		takeProfit.Qty = order.Qty
		takeProfit.ID = utils.GetRandomString(6)
		takeProfit.Type = apiConstants.LimitType
		takeProfit.LimitPrice = order.TakeProfit
		takeProfit.ParentID = &order.ID
		takeProfit.Instrument = order.Instrument
		takeProfit.Side = bracketOrdersSide

		s.orders = append(s.orders, &takeProfit)
	}

	if order.StopLoss != nil {
		stopLoss := api.Order{}
		stopLoss.Qty = order.Qty
		stopLoss.ID = utils.GetRandomString(7)
		stopLoss.Type = apiConstants.StopType
		stopLoss.StopPrice = order.StopLoss
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

// SetOrders ...
func (s *API) SetOrders(orders []*api.Order) {
	s.orders = orders
}

// ModifyOrder ...
func (s *API) ModifyOrder(order *api.Order) (err error) {
	return nil
}

// ClosePosition ...
func (s *API) ClosePosition(marketName string) (err error) {
	if len(s.positions) == 0 {
		return
	}

	positionIndex := 0
	for index, position := range s.positions {
		if position.Instrument == marketName {
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
		if order.Instrument != marketName {
			continue
		}
		if order.LimitPrice != nil {
			takeProfitOrderIndex = index
			hasTakeProfit = true
		}
	}

	if hasTakeProfit {
		s.orders = append(s.orders[:takeProfitOrderIndex], s.orders[takeProfitOrderIndex+1:]...)
	}

	for index, order := range s.orders {
		if order.Instrument != marketName {
			continue
		}
		if order.StopPrice != nil {
			stopLossOrderIndex = index
			hasStopLoss = true
		}
	}

	if hasStopLoss {
		s.orders = append(s.orders[:stopLossOrderIndex], s.orders[stopLossOrderIndex+1:]...)
	}

	return nil
}

func (s *API) AddTrade(
	order *api.Order,
	position *api.Position,
	slippageFunc func(price float64, order *api.Order) float64,
	eurExchangeRate float64,
	lastCandle *types.Candle,
	marketData *types.MarketData,
) {
	finalPrice := float64(.0)

	if order != nil {
		if s.IsStopOrder(order) {
			finalPrice = *order.StopPrice
		}
		if s.IsLimitOrder(order) {
			finalPrice = *order.LimitPrice
		}
	} else {
		finalPrice = lastCandle.Close
	}

	finalPrice = slippageFunc(finalPrice, order)
	tradeResult := (position.AvgPrice - finalPrice) * position.Qty

	if s.IsLongPosition(position) {
		tradeResult = -tradeResult
	}

	tradeResult = adjustResultWithRollover(tradeResult, position, lastCandle, marketData)
	tradeResult = adjustResultWithCommissions(tradeResult, float64(35)/float64(100)/float64(10000), position.AvgPrice, finalPrice, position.Qty)

	tradeResult = tradeResult * eurExchangeRate

	s.state.Balance = s.state.Balance + tradeResult
	s.state.Equity = s.state.Balance

	var side string = "long"
	if s.IsShortPosition(position) {
		side = "short"
	}
	s.trades = append(s.trades, &api.Trade{
		Side:         side,
		Size:         position.Qty,
		InitialPrice: position.AvgPrice,
		FinalPrice:   finalPrice,
		Result:       tradeResult,
		OpenedAt:     time.Unix(*position.CreatedAt, 0),
		ClosedAt:     time.Unix(lastCandle.Timestamp, 0),
	})
}

func (s *API) GetTrades() []*api.Trade {
	return s.trades
}
func (s *API) SetTrades(t []*api.Trade) {
	s.trades = t
}

// CloseOrder ...
func (s *API) CloseOrder(orderID string) (err error) {
	orderIndex := 0
	found := false
	for index, order := range s.orders {
		if order.ID == orderID {
			orderIndex = index
			found = true
			break
		}
	}
	if !found {
		return errors.New("order not found")
	}

	s.orders = append(s.orders[:orderIndex], s.orders[orderIndex+1:]...)
	return nil
}

// GetPositions ...
func (s *API) GetPositions() (positions []*api.Position, err error) {
	return s.positions, nil
}

// SetPositions ...
func (s *API) SetPositions(positions []*api.Position) {
	s.positions = positions
}

// GetState ...
func (s *API) GetState() (state *api.State, err error) {
	return s.state, nil
}

// SetState ...
func (s *API) SetState(state *api.State) {
	s.state = state
}

// ModifyPosition ...
func (s *API) ModifyPosition(marketName string, takeProfit *string, stopLoss *string) (err error) {
	var position *api.Position
	for _, p := range s.positions {
		if p.Instrument != marketName {
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
		if o.Instrument != marketName {
			continue
		}

		if s.IsLimitOrder(o) {
			var aux float64
			var v *float64
			hasTP = true
			if *takeProfit != "" {
				aux = utils.StringToFloat(*takeProfit)
			}
			v = &aux
			o.LimitPrice = v
		}
		if s.IsStopOrder(o) {
			var aux float64
			var v *float64
			hasSL = true
			if *stopLoss != "" {
				aux = utils.StringToFloat(*stopLoss)
			}
			v = &aux
			o.StopPrice = v
		}
	}

	var side string
	if s.IsLongPosition(position) {
		side = apiConstants.ShortSide
	} else {
		side = apiConstants.LongSide
	}
	if !hasTP {
		tpOrder := api.Order{}
		tpOrder.Qty = position.Qty
		tpOrder.ID = utils.GetRandomString(6)
		tpOrder.Type = apiConstants.LimitType
		tp := utils.StringToFloat(*takeProfit)
		tpOrder.LimitPrice = &tp
		tpOrder.Instrument = position.Instrument
		tpOrder.Side = side
		s.CreateOrder(&tpOrder)
	}

	if !hasSL {
		slOrder := api.Order{}
		slOrder.Qty = position.Qty
		slOrder.ID = utils.GetRandomString(7)
		slOrder.Type = apiConstants.StopType
		tp := utils.StringToFloat(*stopLoss)
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

func (s *API) GetBracketOrders(marketName string) (
	slOrder *api.Order,
	tpOrder *api.Order,
) {
	for _, order := range s.orders {
		if order.Instrument != marketName {
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

func (s *API) GetWorkingOrderWithBracketOrders(side string, marketName string, orders []*api.Order) []*api.Order {
	var workingOrders []*api.Order

	for _, order := range s.orders {
		if order.Side != side || order.Instrument != marketName || order.ParentID != nil {
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

// todo: this is same method as in ibroker api
// maybe create a base api class to contain this common logic between api instances
func (s *API) GetSLAndTPOrders(parentID string, orders []*api.Order) (*api.Order, *api.Order) {
	var slOrder *api.Order
	var tpOrder *api.Order
	for _, workingOrder := range orders {
		if workingOrder.ParentID == nil || *workingOrder.ParentID != parentID {
			continue
		}

		if s.IsLimitOrder(workingOrder) {
			tpOrder = workingOrder
		}
		if s.IsStopOrder(workingOrder) {
			slOrder = workingOrder
		}
	}

	return slOrder, tpOrder
}

// CreateAPIServiceInstance ...
func CreateAPIServiceInstance(
	credentials *api.Credentials,
	httpClient httpclient.Interface,
	logger logger.Interface,
) api.Interface {
	instance := &API{
		logger: logger,
	}

	return instance
}

func adjustResultWithRollover(
	tradeResult float64,
	position *api.Position,
	lastCandle *types.Candle,
	marketData *types.MarketData,
) float64 {
	days := int64((lastCandle.Timestamp - *position.CreatedAt) / 60 / 60 / 24)
	if days > 0 {
		days--
	}
	rollover := (marketData.Rollover * position.Qty) / math.Pow(10, float64(marketData.PriceDecimals)-1)
	return tradeResult - float64(days)*rollover
}

func adjustResultWithCommissions(tradeResult float64, commissions, startPrice, finalPrice, size float64) float64 {
	return tradeResult - (commissions*startPrice*size + commissions*finalPrice*size)
}
