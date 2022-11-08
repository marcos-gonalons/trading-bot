package trends

import (
	"TradingBot/src/types"
)

type Service struct{}

func (s *Service) IsBullishTrend(
	candlesAmountToCheck int,
	priceGap float64,
	candles []*types.Candle,
) bool {
	lastCompletedCandleIndex := len(candles) - 1
	lowestValue := candles[lastCompletedCandleIndex].Low
	for i := lastCompletedCandleIndex; i > lastCompletedCandleIndex-candlesAmountToCheck; i-- {
		if i < 1 {
			break
		}
		if candles[i].Low < lowestValue {
			lowestValue = candles[i].Low
		}
	}

	return candles[lastCompletedCandleIndex].Low-lowestValue >= priceGap
}

func (s *Service) IsBearishTrend(
	candlesAmountToCheck int,
	priceGap float64,
	candles []*types.Candle,
) bool {
	lastCompletedCandleIndex := len(candles) - 1
	highestValue := candles[lastCompletedCandleIndex].High
	for i := lastCompletedCandleIndex; i > lastCompletedCandleIndex-candlesAmountToCheck; i-- {
		if i < 1 {
			break
		}
		if candles[i].High > highestValue {
			highestValue = candles[i].High
		}
	}

	return highestValue-candles[lastCompletedCandleIndex].High >= priceGap
}

func GetServiceInstance() Interface {
	return &Service{}
}
