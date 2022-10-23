package brokerSim

import (
	"TradingBot/src/markets"
	"TradingBot/src/services/api"
	"TradingBot/src/types"
)

func OnNewCandle(
	APIData api.DataInterface,
	simulatorAPI api.Interface,
	market markets.MarketInterface,
) {
	orders, _ := simulatorAPI.GetOrders()
	positions, _ := simulatorAPI.GetPositions()
	state, _ := simulatorAPI.GetState()

	///////////////////////
	spread := float32(.0)
	stopOrderSlippage := float32(.0)
	///////////////////////

	candles := market.GetCandlesHandler().GetCandles()
	orderIDsToRemove := []string{}
	for _, order := range orders {
		lastCandle := candles[len(candles)-2]

		if simulatorAPI.IsMarketOrder(order) {
			position := findPosition(positions, order.Instrument)
			if position != nil {
				panic("the strategies must never create a market order if there is already an open position")
			}
			positions = append(positions, createNewPosition(float32(lastCandle.Open), order, order.Qty, stopOrderSlippage, lastCandle))
			simulatorAPI.SetPositions(positions)

			market.SetCurrentPositionExecutedAt(lastCandle.Timestamp)
			simulatorAPI.CloseOrder(order.ID)
			continue
		}

		orderExecutionPrice := getOrderExecutionPrice(simulatorAPI, order, spread)
		positionPrice := float32(0)

		if isPriceWithinCandle(float64(orderExecutionPrice), lastCandle) {
			positionPrice = orderExecutionPrice
		}
		if hasCandleGapOvercameExecutionPrice(float64(orderExecutionPrice), lastCandle, candles[len(candles)-3]) {
			positionPrice = float32(lastCandle.Open)
		}

		if positionPrice == 0 {
			continue
		}

		position := findPosition(positions, order.Instrument)
		if position == nil {
			if order.ParentID != nil {
				continue
			}
			positions = append(positions, createNewPosition(positionPrice, order, order.Qty, stopOrderSlippage, lastCandle))
			simulatorAPI.SetPositions(positions)

			market.SetCurrentPositionExecutedAt(lastCandle.Timestamp)
		} else {
			if position.Side == order.Side {
				position.Qty = position.Qty + order.Qty
				position.AvgPrice = (position.AvgPrice + positionPrice) / 2

				panic("right now it will never enter here if everything works as expected")
			} else {
				simulatorAPI.AddTrade(
					order,
					position,
					func(price float32, order *api.Order) float32 {
						return addSlippage(price, order, stopOrderSlippage)
					},
					market.GetMarketData().EurExchangeRate,
					lastCandle,
					market.GetMarketData(),
				)

				if order.Qty == position.Qty {
					simulatorAPI.ClosePosition(position.Instrument)
				} else {
					if order.Qty > position.Qty {
						simulatorAPI.ClosePosition(position.Instrument)
						p, _ := simulatorAPI.GetPositions()
						positions = append(p, createNewPosition(positionPrice, order, order.Qty-position.Qty, stopOrderSlippage, lastCandle))
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

	p, _ := simulatorAPI.GetPositions()
	APIData.SetPositions(p)

	APIData.SetState(state)
	simulatorAPI.SetState(state)
}

func isPriceWithinCandle(price float64, candle *types.Candle) bool {
	return price >= candle.Low && price <= candle.High
}

func getOrderExecutionPrice(
	simulatorAPI api.Interface,
	order *api.Order,
	spread float32,
) float32 {
	if simulatorAPI.IsLimitOrder(order) {
		if simulatorAPI.IsLongOrder(order) {
			return *order.LimitPrice - spread/2
		}
		if simulatorAPI.IsShortOrder(order) {
			return *order.LimitPrice + spread/2
		}
	}

	if simulatorAPI.IsStopOrder(order) {
		if simulatorAPI.IsLongOrder(order) {
			return *order.StopPrice + spread/2
		}
		if simulatorAPI.IsShortOrder(order) {
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
	lastCandle *types.Candle,
) *api.Position {
	price := addSlippage(positionPrice, order, stopOrderSlippage)

	return &api.Position{
		Instrument:   order.Instrument,
		Qty:          size,
		Side:         order.Side,
		AvgPrice:     price,
		UnrealizedPl: .0,
		CreatedAt:    &lastCandle.Timestamp,
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
