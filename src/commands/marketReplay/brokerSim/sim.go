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

	candles := market.GetCandlesHandler().GetCandles()
	orderIDsToRemove := []string{}
	for _, order := range orders {
		lastCandle := candles[len(candles)-2]

		if simulatorAPI.IsMarketOrder(order) {
			position := findPosition(positions, order.Instrument)
			if position != nil {
				panic("the strategies must never create a market order if there is already an open position")
			}
			positions = append(positions, createNewPosition(lastCandle.Open, order, order.Qty, market.GetMarketData().SimulatorData.Spread/2, lastCandle, simulatorAPI))
			simulatorAPI.SetPositions(positions)

			market.SetCurrentPositionExecutedAt(lastCandle.Timestamp)
			orderIDsToRemove = append(orderIDsToRemove, order.ID)
			continue
		}

		orderExecutionPrice := getOrderExecutionPrice(simulatorAPI, order, market.GetMarketData().SimulatorData.Spread)
		price := float64(0)

		if isPriceWithinCandle(orderExecutionPrice, lastCandle) {
			price = orderExecutionPrice
		}
		if hasCandleGapOvercameExecutionPrice(orderExecutionPrice, lastCandle, candles[len(candles)-3]) {
			price = lastCandle.Open
		}

		if price == 0 {
			continue
		}

		position := findPosition(positions, order.Instrument)
		if position == nil {
			if order.ParentID != nil { // If the price reaches the SL or TP order, there is nothing to do since there isn't an open position.
				continue
			}

			// Here it means that we triggered a limit or a stop order, and there wasn't any open position
			// So we need to create the position
			positions = append(positions, createNewPosition(price, order, order.Qty, market.GetMarketData().SimulatorData.Spread/2, lastCandle, simulatorAPI))
			simulatorAPI.SetPositions(positions)

			market.SetCurrentPositionExecutedAt(lastCandle.Timestamp)
		} else {
			if position.Side == order.Side {
				position.Qty = position.Qty + order.Qty
				position.AvgPrice = (position.AvgPrice + price) / 2

				panic("right now it will never enter here if everything works as expected")
			} else {
				// This is a SL or TP order.
				simulatorAPI.AddTrade(
					order,
					position,
					func(price float64, order *api.Order) float64 {
						return addSlippage(price, order, market.GetMarketData().SimulatorData.Slippage, simulatorAPI)
					},
					market.GetMarketData().EurExchangeRate,
					lastCandle,
					market.GetMarketData(),
				)

				if order.Qty == position.Qty {
					simulatorAPI.ClosePosition(position.Instrument)
					break
				} else {
					if order.Qty > position.Qty {
						simulatorAPI.ClosePosition(position.Instrument)
						p, _ := simulatorAPI.GetPositions()
						positions = append(p, createNewPosition(price, order, order.Qty-position.Qty, market.GetMarketData().SimulatorData.Slippage, lastCandle, simulatorAPI))
						simulatorAPI.SetPositions(positions)
						break
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
	spread float64,
) float64 {
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
	positionPrice float64,
	order *api.Order,
	size float64,
	slippage float64,
	lastCandle *types.Candle,
	simulatorAPI api.Interface,
) *api.Position {
	price := addSlippage(positionPrice, order, slippage, simulatorAPI)

	return &api.Position{
		Instrument:   order.Instrument,
		Qty:          size,
		Side:         order.Side,
		AvgPrice:     price,
		UnrealizedPl: .0,
		CreatedAt:    &lastCandle.Timestamp,
	}
}

func addSlippage(price float64, order *api.Order, slippage float64, simulatorAPI api.Interface) float64 {
	if simulatorAPI.IsLimitOrder(order) {
		return price
	}

	if simulatorAPI.IsLongOrder(order) {
		return price + slippage
	}
	if simulatorAPI.IsShortOrder(order) {
		return price - slippage
	}

	return price
}
