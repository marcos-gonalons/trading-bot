package horizontalLevels

import (
	"TradingBot/src/types"
	"errors"
)

type Service struct{}

func (s *Service) GetResistancePrice(
	candlesWithLowerPriceToBeConsideredTop types.CandlesAmountForHorizontalLevel,
	lastCompletedCandleIndex int,
	candles []*types.Candle,
) (price float64, err error) {
	return s.getPrice(
		candlesWithLowerPriceToBeConsideredTop,
		lastCompletedCandleIndex,
		ResistanceName,
		candles,
	)
}

func (s *Service) GetSupportPrice(
	candlesWithHigherPriceToBeConsideredBottom types.CandlesAmountForHorizontalLevel,
	lastCompletedCandleIndex int,
	candles []*types.Candle,
) (price float64, err error) {
	return s.getPrice(
		candlesWithHigherPriceToBeConsideredBottom,
		lastCompletedCandleIndex,
		SupportName,
		candles,
	)
}

func (s *Service) getPrice(
	candlesAmount types.CandlesAmountForHorizontalLevel,
	lastCompletedCandleIndex int,
	supportOrResistance string,
	candles []*types.Candle,
) (float64, error) {
	candlesToCheck := 300
	lastCompletedCandleIndex++
	for x := lastCompletedCandleIndex; x > lastCompletedCandleIndex-candlesToCheck; x-- {
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
			if j < 1 || j > lastCompletedCandleIndex {
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
			return candles[horizontalLevelCandleIndex].High, nil
		}
		if supportOrResistance == SupportName {
			return candles[horizontalLevelCandleIndex].Low, nil
		}

	}

	return 0, errors.New("unable to find the horizontal level")
}

func GetServiceInstance() Interface {
	return &Service{}
}
