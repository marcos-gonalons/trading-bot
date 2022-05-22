package brokerSim

import (
	"TradingBot/src/services/api"
	"TradingBot/src/strategies/markets/interfaces"
	"TradingBot/src/types"
)

type Trade struct{}

type BrokerSim struct{}

var eurExchangeRate float64
var trades []*Trade

func (s *BrokerSim) OnNewCandle(
	APIData api.DataInterface,
	simulatorAPI api.Interface,
	strat interfaces.MarketInterface,
) {
	eurExchangeRate = strat.Parent().GetEurExchangeRate()

	orders, _ := simulatorAPI.GetOrders()
	positions, _ := simulatorAPI.GetPositions()
	state, _ := simulatorAPI.GetState()

	///////////////////////
	spread := float32(.0)
	stopOrderSlippage := float32(.0)
	///////////////////////

	candles := strat.Parent().GetCandlesHandler().GetCandles()
	orderIDsToRemove := []string{}
	for _, order := range orders {
		lastCandle := candles[len(candles)-1]

		orderExecutionPrice := getOrderExecutionPrice(order, spread)
		positionPrice := float32(0)

		if isPriceWithinCandle(float64(orderExecutionPrice), lastCandle) {
			positionPrice = orderExecutionPrice
		}
		if hasCandleGapOvercameExecutionPrice(float64(orderExecutionPrice), lastCandle, candles[len(candles)-2]) {
			positionPrice = float32(lastCandle.Open)
		}

		if positionPrice == 0 {
			continue
		}

		position := findPosition(positions, order.Instrument)
		if position == nil {
			positions = append(positions, createNewPosition(positionPrice, order, order.Qty, stopOrderSlippage))
			APIData.SetPositions(positions)
			simulatorAPI.SetPositions(positions)
		} else {
			if position.Side == order.Side {
				position.Qty = position.Qty + order.Qty
				position.AvgPrice = (position.AvgPrice + positionPrice) / 2

				panic("right now it will never enter here if everything works as expected")
			} else {
				addTrade(position, order, state, order.Qty, stopOrderSlippage)

				if order.Qty == position.Qty {
					simulatorAPI.ClosePosition(position.Instrument)
				} else {
					if order.Qty > position.Qty {
						simulatorAPI.ClosePosition(position.Instrument)
						positions = append(positions, createNewPosition(positionPrice, order, order.Qty-position.Qty, stopOrderSlippage))
						APIData.SetPositions(positions)
						simulatorAPI.SetPositions(positions)
					} else {
						position.Qty -= order.Qty
					}
					panic("should never enter here if everything works as expected")
				}
			}
		}

		orderIDsToRemove = append(orderIDsToRemove, order.ID)
	}

	for _, orderID := range orderIDsToRemove {
		simulatorAPI.CloseOrder(orderID)
	}

	o, _ := simulatorAPI.GetOrders()
	APIData.SetOrders(o)

	APIData.SetState(state)
	simulatorAPI.SetState(state)
}

func isPriceWithinCandle(price float64, candle *types.Candle) bool {
	return price >= candle.Low && price <= candle.High
}

func getOrderExecutionPrice(order *api.Order, spread float32) float32 {
	// todo: simulator constants, like ibroker constants
	if order.Type == "limit" {
		if order.Side == "buy" {
			return *order.LimitPrice - spread/2
		}
		if order.Side == "sell" {
			return *order.LimitPrice + spread/2
		}
	}

	if order.Type == "stop" {
		if order.Side == "buy" {
			return *order.StopPrice + spread/2
		}
		if order.Side == "sell" {
			return *order.StopPrice - spread/2
		}
	}

	panic("invalid order")
}

func hasCandleGapOvercameExecutionPrice(
	price float64,
	currentCandle, previousCandle *types.Candle,
) bool {
	if previousCandle != nil {
		return (currentCandle.Low > price && previousCandle.High < price) ||
			(currentCandle.High < price && previousCandle.Low > price)
	}
	return false
}

func findPosition(positions []*api.Position, instrument string) *api.Position {
	for _, pos := range positions {
		if pos.Instrument == instrument {
			return pos
		}
	}
	return nil
}

func createNewPosition(
	positionPrice float32,
	order *api.Order,
	size float32,
	stopOrderSlippage float32,
) *api.Position {
	price := addSlippage(positionPrice, order, stopOrderSlippage)

	return &api.Position{
		Instrument:   order.Instrument,
		Qty:          size,
		Side:         order.Side,
		AvgPrice:     price,
		UnrealizedPl: .0,
	}
}

func addSlippage(price float32, order *api.Order, slippage float32) float32 {
	if order.Type != "stop" {
		return price
	}

	if order.Side == "buy" {
		return price + slippage
	}
	if order.Side == "sell" {
		return price - slippage
	}

	return price
}

func addTrade(
	position *api.Position,
	order *api.Order,
	state *api.State,
	size float32,
	stopOrderSlippage float32,
) {
	finalPrice := float32(.0)
	if order.Type == "stop" {
		finalPrice = *order.StopPrice
	}
	if order.Type == "limit" {
		finalPrice = *order.LimitPrice
	}
	finalPrice = addSlippage(finalPrice, order, stopOrderSlippage)
	tradeResult := position.AvgPrice - finalPrice
	if position.Side == "buy" {
		state.Balance = state.Balance - float64(tradeResult)
	}
	if position.Side == "sell" {
		state.Balance = state.Balance + float64(tradeResult)
	}

	state.Balance *= float64(size)
	state.Balance *= eurExchangeRate

	state.Equity = state.Balance

}
