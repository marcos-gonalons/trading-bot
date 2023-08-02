package ranges

import (
	"TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"errors"
)

func getOrderPrice(
	params types.MarketStrategyParams,
	resistancesAverage float64,
	supportsAverage float64,
	lastCompletedCandle *types.Candle,
	longOrShort string,
) (float64, error) {
	switch params.Ranges.OrderType {
	case constants.LimitType:
		if longOrShort == constants.LongSide {
			price := supportsAverage + float64(params.Ranges.PriceOffset)
			if lastCompletedCandle.Close <= price {
				return price, errors.New("order is limit but last completed candle close is below the price")
			}
			return price, nil
		} else {
			price := resistancesAverage + float64(params.Ranges.PriceOffset)
			if lastCompletedCandle.Close >= price {
				return price, errors.New("order is limit but last completed candle close is below the price")
			}
			return price, nil
		}
	case constants.StopType:
		if longOrShort == constants.LongSide {
			price := resistancesAverage + float64(params.Ranges.PriceOffset)
			if price <= lastCompletedCandle.Close {
				return price, errors.New("order is stop but the price is below the last completed candle close")
			}
			return price, nil
		} else {
			price := supportsAverage + float64(params.Ranges.PriceOffset)
			if price >= lastCompletedCandle.Close {
				return price, errors.New("order is stop but the price is below the last completed candle close")
			}
			return price, nil
		}
	case constants.MarketType:
		return lastCompletedCandle.Close, nil
	}

	return 0, nil
}

func getStopLoss(
	params types.MarketStrategyParams,
	resistancesAverage float64,
	supportsAverage float64,
	orderPrice float64,
	longOrShort string,
) (sl float64) {
	switch params.Ranges.StopLossStrategy {
	case "half":
		sl = (resistancesAverage + supportsAverage) / 2
		break
	case "level":
	case "level-with-offset":
		if params.Ranges.OrderType == constants.StopType {
			if longOrShort == constants.LongSide {
				sl = supportsAverage
			} else {
				sl = resistancesAverage
			}
		}
		if params.Ranges.OrderType == constants.LimitType {
			if longOrShort == constants.LongSide {
				sl = resistancesAverage
			} else {
				sl = supportsAverage
			}
		}

		if params.Ranges.StopLossStrategy == "level" {
			break
		}

		if longOrShort == constants.LongSide {
			sl = sl - params.StopLossDistance
		} else {
			sl = sl + params.StopLossDistance
		}
		break
	case "distance":
		if longOrShort == constants.LongSide {
			sl = orderPrice - params.StopLossDistance
		} else {
			sl = orderPrice + params.StopLossDistance
		}
		break
	default:
		panic("Invalid stop loss strategy -> " + params.Ranges.StopLossStrategy)
	}

	if longOrShort == constants.LongSide {
		if orderPrice-sl > params.MaxStopLossDistance {
			sl = orderPrice - params.MaxStopLossDistance
		}
	} else {
		if sl-orderPrice > params.MaxStopLossDistance {
			sl = orderPrice + params.MaxStopLossDistance
		}
	}

	return
}

func getTakeProfit(
	params types.MarketStrategyParams,
	resistancesAverage float64,
	supportsAverage float64,
	orderPrice float64,
	longOrShort string,
) float64 {
	switch params.Ranges.TakeProfitStrategy {
	case "half":
		return (resistancesAverage + supportsAverage) / 2
	case "level":
		if longOrShort == constants.LongSide {
			return resistancesAverage
		} else {
			return supportsAverage
		}
	case "level-with-offset":
		if longOrShort == constants.LongSide {
			return resistancesAverage + params.TakeProfitDistance
		} else {
			return supportsAverage - params.TakeProfitDistance
		}
	case "distance":
		if longOrShort == constants.LongSide {
			return orderPrice + params.TakeProfitDistance
		} else {
			return orderPrice - params.TakeProfitDistance
		}
	}
	panic("Invalid take profit startegy -> " + utils.GetStringRepresentation(params.Ranges.TakeProfitStrategy))
}
