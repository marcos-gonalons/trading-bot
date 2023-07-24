package horizontalLevels

import (
	"errors"
)

type Service struct{}

func (s *Service) GetResistance(params GetLevelParams) (*Level, error) {
	return s.getLevel(RESISTANCE_TYPE, params)
}
func (s *Service) GetSupport(params GetLevelParams) (*Level, error) {
	return s.getLevel(SUPPORT_TYPE, params)
}

func (s *Service) getLevel(levelType LevelType, params GetLevelParams) (*Level, error) {
	for x := params.StartAt; x > params.StartAt-params.CandlesToCheck; x-- {
		if x < 0 {
			break
		}

		if !s.isLevelValid(x, levelType, params) {
			continue
		}

		return &Level{
			Type:        levelType,
			CandleIndex: x,
			Candle:      params.Candles[x],
		}, nil
	}

	return nil, errors.New("Unable to find the horizontal level")
}

func (s *Service) isLevelValid(index int64, levelType LevelType, params GetLevelParams) bool {
	totalCandles := int64(len(params.Candles))

	if index >= totalCandles || index < 0 {
		return false
	}

	// Future candles
	for i := index + 1; i < index+1+int64(params.CandlesAmountToBeConsideredHorizontalLevel.Future); i++ {
		if i == totalCandles {
			return false
		}
		if levelType == RESISTANCE_TYPE {
			if params.Candles[i].High > params.Candles[index].High {
				return false
			}
		}
		if levelType == SUPPORT_TYPE {
			if params.Candles[i].Low < params.Candles[index].Low {
				return false
			}
		}
	}

	// Past candles
	for i := index - int64(params.CandlesAmountToBeConsideredHorizontalLevel.Past); i < index; i++ {
		if i < 0 {
			return false
		}
		if levelType == RESISTANCE_TYPE {
			if params.Candles[i].High > params.Candles[index].High {
				return false
			}
		}

		if levelType == SUPPORT_TYPE {
			if params.Candles[i].Low < params.Candles[index].Low {
				return false
			}
		}
	}

	return true
}

func GetServiceInstance() Interface {
	return &Service{}
}
