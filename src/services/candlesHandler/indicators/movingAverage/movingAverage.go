package movingAverage

import (
	"TradingBot/src/types"

	"github.com/thoas/go-funk"
)

type MovingAverage struct{}

const EMA_SMOOTHING_FACTOR = 2

func (s *MovingAverage) AddData(candles []*types.Candle, lastCandleOnly bool) {
	emaLengths := []int64{9, 21, 200}

	start := 0
	if lastCandleOnly {
		start = len(candles) - 1
	}

	for i := start; i < len(candles); i++ {
		movingAverages := []types.MovingAverage{}
		for _, emaLength := range emaLengths {
			movingAverages = append(movingAverages, types.MovingAverage{
				Description:   "Exponential Moving Average",
				Name:          "EMA",
				Value:         s.getExponentialMovingAverage(emaLength, candles, i),
				CandlesAmount: emaLength,
			})

			candles[i].Indicators.MovingAverages = movingAverages
		}
	}
}

func (s *MovingAverage) getExponentialMovingAverage(
	candlesAmount int64,
	candles []*types.Candle,
	index int,
) float64 {
	if index == 0 {
		return candles[0].Close
	}

	multiplier := float64(EMA_SMOOTHING_FACTOR) / float64((candlesAmount + 1))
	previousEMA := funk.Find(
		candles[index-1].Indicators.MovingAverages,
		func(ma types.MovingAverage) bool {
			return ma.Name == "EMA" && ma.CandlesAmount == candlesAmount
		},
	)

	previousValue := previousEMA.(types.MovingAverage).Value

	return candles[index].Close*multiplier + (previousValue * (1 - multiplier))
}
