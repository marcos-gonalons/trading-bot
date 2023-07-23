package horizontalLevels

import (
	"TradingBot/src/types"
	"errors"
)

type Service struct{}

func (s *Service) GetResistancePrice(
	candlesWithLowerPriceToBeConsideredTop types.CandlesAmountForHorizontalLevel,
	candles []*types.Candle,
	startingIndex int,
) (price float64, foundAtIndex int, err error) {
	return s.getPrice(
		candlesWithLowerPriceToBeConsideredTop,
		RESISTANCE,
		candles,
		startingIndex,
	)
}

func (s *Service) GetSupportPrice(
	candlesWithHigherPriceToBeConsideredBottom types.CandlesAmountForHorizontalLevel,
	candles []*types.Candle,
	startingIndex int,
) (price float64, foundAtIndex int, err error) {
	return s.getPrice(
		candlesWithHigherPriceToBeConsideredBottom,
		SUPPORT,
		candles,
		startingIndex,
	)
}

// todo: call this getPriceOLD and keep using it for emacrossover strategy
// until I find a better strategy that uses the correct method to get the horizontal level
func (s *Service) getPrice(
	candlesAmount types.CandlesAmountForHorizontalLevel,
	levelType LevelType,
	candles []*types.Candle,
	startingIndex int,
) (float64, int, error) {
	candlesToCheck := 300

	for x := startingIndex; x > startingIndex-candlesToCheck; x-- {
		if x < 0 {
			break
		}

		horizontalLevelCandleIndex := x - candlesAmount.Future
		if horizontalLevelCandleIndex < 0 || x < candlesAmount.Future+candlesAmount.Past {
			continue
		}

		futureCandlesOvercomePrice := false
		for j := horizontalLevelCandleIndex + 1; j < x; j++ {
			if levelType == RESISTANCE {
				if candles[j].High > candles[horizontalLevelCandleIndex].High {
					futureCandlesOvercomePrice = true
					break
				}
			}
			if levelType == SUPPORT {
				if candles[j].Low < candles[horizontalLevelCandleIndex].Low {
					futureCandlesOvercomePrice = true
					break
				}
			}
		}

		if futureCandlesOvercomePrice {
			continue
		}

		pastCandlesOvercomePrice := false
		for j := horizontalLevelCandleIndex - candlesAmount.Past; j < horizontalLevelCandleIndex; j++ {
			if j < 1 || j > startingIndex {
				continue
			}
			if levelType == RESISTANCE {
				if candles[j].High > candles[horizontalLevelCandleIndex].High {
					pastCandlesOvercomePrice = true
					break
				}
			}
			if levelType == SUPPORT {
				if candles[j].Low < candles[horizontalLevelCandleIndex].Low {
					pastCandlesOvercomePrice = true
					break
				}
			}
		}

		if pastCandlesOvercomePrice {
			continue
		}

		if levelType == RESISTANCE {
			return candles[horizontalLevelCandleIndex].High, horizontalLevelCandleIndex, nil
		}
		if levelType == SUPPORT {
			return candles[horizontalLevelCandleIndex].Low, horizontalLevelCandleIndex, nil
		}

	}

	return 0, 0, errors.New("unable to find the horizontal level")
}

func GetServiceInstance() Interface {
	return &Service{}
}
