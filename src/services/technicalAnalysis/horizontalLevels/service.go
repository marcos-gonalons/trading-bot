package horizontalLevels

import (
	"TradingBot/src/services/candlesHandler"
	"errors"
)

type Service struct {
	CandlesHandler candlesHandler.Interface
}

// ResistanceName ...
const ResistanceName = "resistance"

// SupportName ...
const SupportName = "support"

// FutureCandlesOvercamePriceError ...
const FutureCandlesOvercamePriceError = "future_overcame"

// PastCandlesOvercamePriceError ...
const PastCandlesOvercamePriceError = "past_overcame"

// NotEnoughCandlesError ...
const NotEnoughCandlesError = "not_enough_candles"

func (s *Service) GetResistancePrice(candlesWithLowerPriceToBeConsideredTop int) (price float64, err error) {
	return s.getPrice(candlesWithLowerPriceToBeConsideredTop, ResistanceName)
}

func (s *Service) GetSupportPrice(candlesWithHigherPriceToBeConsideredBottom int) (price float64, err error) {
	return s.getPrice(candlesWithHigherPriceToBeConsideredBottom, SupportName)
}

func (s *Service) getPrice(candlesAmount int, supportOrResistance string) (price float64, err error) {
	candles := s.CandlesHandler.GetCandles()
	lastCandlesIndex := len(candles) - 1

	horizontalLevelCandleIndex := lastCandlesIndex - candlesAmount
	if horizontalLevelCandleIndex < 0 || lastCandlesIndex < candlesAmount*2 {
		err = errors.New(NotEnoughCandlesError)
		return
	}

	futureCandlesOvercomePrice := false
	for j := horizontalLevelCandleIndex + 1; j < lastCandlesIndex-1; j++ {
		if supportOrResistance == ResistanceName {
			if candles[j].High >= candles[horizontalLevelCandleIndex].High {
				futureCandlesOvercomePrice = true
				break
			}
		}
		if supportOrResistance == SupportName {
			if candles[j].Low <= candles[horizontalLevelCandleIndex].Low {
				futureCandlesOvercomePrice = true
				break
			}
		}
	}

	if futureCandlesOvercomePrice {
		err = errors.New(FutureCandlesOvercamePriceError)
		return
	}

	pastCandlesOvercomePrice := false
	for j := horizontalLevelCandleIndex - candlesAmount; j < horizontalLevelCandleIndex; j++ {
		if j < 1 || j > lastCandlesIndex {
			continue
		}
		if supportOrResistance == ResistanceName {
			if candles[j].High >= candles[horizontalLevelCandleIndex].High {
				pastCandlesOvercomePrice = true
				break
			}
		}
		if supportOrResistance == SupportName {
			if candles[j].Low <= candles[horizontalLevelCandleIndex].Low {
				pastCandlesOvercomePrice = true
				break
			}
		}
	}

	if pastCandlesOvercomePrice {
		err = errors.New(PastCandlesOvercamePriceError)
		return
	}

	if supportOrResistance == ResistanceName {
		price = candles[horizontalLevelCandleIndex].High
	}
	if supportOrResistance == SupportName {
		price = candles[horizontalLevelCandleIndex].Low
	}

	return
}

func GetServiceInstance(candlesHandler candlesHandler.Interface) Interface {
	return &Service{
		CandlesHandler: candlesHandler,
	}
}
