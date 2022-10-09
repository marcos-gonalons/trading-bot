package movingAverage

import (
	"TradingBot/src/types"

	"github.com/thoas/go-funk"
)

type MovingAverage struct{}

const EMA_SMOOTHING_FACTOR = 2

func (s *MovingAverage) AddData(candles []*types.Candle) {
	emaLengths := []int64{9, 21, 200}

	for i, candle := range candles {
		if i == 0 {
			continue
		}

		movingAverages := []types.MovingAverage{}
		for _, emaLength := range emaLengths {
			movingAverages = append(movingAverages, types.MovingAverage{
				Description:   "Exponential Moving Average",
				Name:          "EMA",
				Value:         s.getExponentialMovingAverage(emaLength, candles),
				CandlesAmount: emaLength,
			})

			candle.Indicators.MovingAverages = movingAverages
		}
	}
}

func (s *MovingAverage) getExponentialMovingAverage(
	candlesAmount int64,
	candles []*types.Candle,
) float64 {
	multiplier := float64(EMA_SMOOTHING_FACTOR / (candlesAmount + 1))
	previousEMA := funk.Find(
		candles[len(candles)-2].Indicators.MovingAverages,
		func(ma types.MovingAverage) bool {
			return ma.Name == "EMA" && ma.CandlesAmount == candlesAmount
		},
	)

	previousValue := candles[len(candles)-2].Close
	if previousEMA != nil {
		previousValue = previousEMA.(types.MovingAverage).Value
	}

	return candles[len(candles)-1].Close*multiplier + (previousValue + (1 - multiplier))
}
