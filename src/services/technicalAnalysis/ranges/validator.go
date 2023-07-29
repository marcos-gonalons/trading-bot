package ranges

import (
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/types"
	"math"

	"github.com/thoas/go-funk"
)

type IsRangeValidParams struct {
	Range               []*horizontalLevels.Level
	ValidationParams    *types.Ranges
	Candles             []*types.Candle
	LastCompletedCandle *types.Candle
}

func IsRangeValid(params *IsRangeValidParams) bool {
	return (hasEnoughLevels(params.Range) &&
		hasAlternatingLevels(params.Range) &&
		areSameLevelPricesCloseEnough(params.Range, params.ValidationParams.MaxPriceDifferenceForSameHorizontalLevel) &&
		isPriceDifferenceBetweenLevelsHighEnough(params.Range, params.ValidationParams.MinPriceDifferenceBetweenRangePoints) &&
		areAllCandlesBoundedBetweenLevels(params.Range, params.Candles) &&
		hasAppropiateAmountOfCandlesBetweenLevels(params.Range, params.ValidationParams.MinCandlesBetweenRangePoints, params.ValidationParams.MaxCandlesBetweenRangePoints) &&
		isCandleInsideTheRange(params.Range, params.LastCompletedCandle))
}

func hasEnoughLevels(r []*horizontalLevels.Level) bool {
	return len(r) >= 2
}

// Each level of the range must alternate between resistance and support.
func hasAlternatingLevels(r []*horizontalLevels.Level) bool {
	for i := 1; i < len(r); i++ {
		if r[i-1].IsResistance() && r[i].IsResistance() {
			return false
		}
		if r[i-1].IsSupport() && r[i].IsSupport() {
			return false
		}
	}

	return true
}

// Price of the same level must be close to each other.
func areSameLevelPricesCloseEnough(r []*horizontalLevels.Level, maxPriceDifference float64) bool {
	resistances := funk.Filter(r, func(l *horizontalLevels.Level) bool {
		return l.IsResistance()
	}).([]*horizontalLevels.Level)

	supports := funk.Filter(r, func(l *horizontalLevels.Level) bool {
		return l.IsSupport()
	}).([]*horizontalLevels.Level)

	priceToCompare := resistances[0].Candle.High
	for _, level := range resistances {
		if math.Abs(priceToCompare-level.Candle.High) > maxPriceDifference {
			return false
		}
	}

	priceToCompare = supports[0].Candle.Low
	for _, level := range supports {
		if math.Abs(priceToCompare-level.Candle.Low) > maxPriceDifference {
			return false
		}
	}

	return true
}

// Price difference between levels must be high enough.
func isPriceDifferenceBetweenLevelsHighEnough(r []*horizontalLevels.Level, minPriceDifference float64) bool {
	for i := 1; i < len(r); i++ {
		previousLevelPrice := r[i-1].GetPrice()
		currentLevelPrice := r[i].GetPrice()
		if math.Abs(previousLevelPrice-currentLevelPrice) < minPriceDifference {
			return false
		}
	}
	return true
}

// All candles between levels must not be lower or higher than the level.
func areAllCandlesBoundedBetweenLevels(r []*horizontalLevels.Level, candles []*types.Candle) bool {
	for i := 0; i < len(r)-1; i++ {
		higherIndex := r[i].CandleIndex - 1
		lowerIndex := r[i+1].CandleIndex + 1
		if higherIndex <= lowerIndex {
			return false
		}
		for j := higherIndex; j > lowerIndex; j-- {
			if r[i].IsResistance() {
				if candles[j].High > r[i].Candle.High {
					return false
				}
				if candles[j].Low < r[i+1].Candle.Low {
					return false
				}
			}
			if r[i].IsSupport() {
				if candles[j].Low < r[i].Candle.Low {
					return false
				}
				if candles[j].High < r[i+1].Candle.High {
					return false
				}
			}
		}
	}

	return true
}

// There must be enough candles between each level and not too many
func hasAppropiateAmountOfCandlesBetweenLevels(r []*horizontalLevels.Level, minCandles int64, maxCandles int64) bool {
	for i := 0; i < len(r)-1; i++ {
		diff := math.Abs(float64(r[i].CandleIndex) - float64(r[i+1].CandleIndex))
		if diff < float64(minCandles) || diff > float64(maxCandles) {
			return false
		}
	}

	return true
}

// candle.close must be between the lowest resistance price and the highest support price.
func isCandleInsideTheRange(r []*horizontalLevels.Level, candle *types.Candle) bool {
	lowestResistance := math.Inf(1)
	highestSupport := math.Inf(-1)
	for _, level := range r {
		if level.IsResistance() && level.Candle.High < lowestResistance {
			lowestResistance = level.Candle.High
		}

		if level.IsSupport() && level.Candle.Low > highestSupport {
			highestSupport = level.Candle.Low
		}
	}
	if candle.Close > lowestResistance || candle.Close < highestSupport {
		return false
	}

	return true
}
