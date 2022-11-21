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
		ResistanceName,
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
		SupportName,
		candles,
		startingIndex,
	)
}

func (s *Service) getPrice(
	candlesAmount types.CandlesAmountForHorizontalLevel,
	supportOrResistance string,
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
			if supportOrResistance == ResistanceName {
				if candles[j].High > candles[horizontalLevelCandleIndex].High {
					futureCandlesOvercomePrice = true
					break
				}
			}
			if supportOrResistance == SupportName {
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
			if supportOrResistance == ResistanceName {
				if candles[j].High > candles[horizontalLevelCandleIndex].High {
					pastCandlesOvercomePrice = true
					break
				}
			}
			if supportOrResistance == SupportName {
				if candles[j].Low < candles[horizontalLevelCandleIndex].Low {
					pastCandlesOvercomePrice = true
					break
				}
			}
		}

		if pastCandlesOvercomePrice {
			continue
		}

		if supportOrResistance == ResistanceName {
			return candles[horizontalLevelCandleIndex].High, horizontalLevelCandleIndex, nil
		}
		if supportOrResistance == SupportName {
			return candles[horizontalLevelCandleIndex].Low, horizontalLevelCandleIndex, nil
		}

	}

	return 0, 0, errors.New("unable to find the horizontal level")
}

func GetServiceInstance() Interface {
	return &Service{}
}
