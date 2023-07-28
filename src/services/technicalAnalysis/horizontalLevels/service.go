package horizontalLevels

import "TradingBot/src/types"

type Service struct{}

func (s *Service) GetLevel(levelType types.LevelType, params GetLevelParams) *Level {
	if levelType == types.RESISTANCE_TYPE {
		return s.GetResistance(params)
	}
	if levelType == types.SUPPORT_TYPE {
		return s.GetSupport(params)
	}
	panic("Invalid level type")
}

func (s *Service) GetResistance(params GetLevelParams) *Level {
	return s.getLevel(types.RESISTANCE_TYPE, params)
}
func (s *Service) GetSupport(params GetLevelParams) *Level {
	return s.getLevel(types.SUPPORT_TYPE, params)
}

func (s *Service) getLevel(levelType types.LevelType, params GetLevelParams) *Level {
	for x := params.StartAt; x > params.StartAt-params.CandlesToCheck; x-- {
		if x < 0 {
			break
		}

		level := &Level{
			Type:        levelType,
			CandleIndex: x,
			Candle:      params.Candles[x],
		}

		if !s.isLevelValid(level, params) {
			continue
		}

		return level
	}

	return nil
}

func (s *Service) isLevelValid(level *Level, params GetLevelParams) bool {
	totalCandles := int64(len(params.Candles))

	if level.CandleIndex >= totalCandles || level.CandleIndex < 0 {
		return false
	}

	// Future candles
	for i := level.CandleIndex + 1; i < level.CandleIndex+1+int64(params.CandlesAmountToBeConsideredHorizontalLevel.Future); i++ {
		if i == totalCandles {
			return false
		}
		if level.IsResistance() {
			if params.Candles[i].High > params.Candles[level.CandleIndex].High {
				return false
			}
		}
		if level.IsSupport() {
			if params.Candles[i].Low < params.Candles[level.CandleIndex].Low {
				return false
			}
		}
	}

	// Past candles
	for i := level.CandleIndex - int64(params.CandlesAmountToBeConsideredHorizontalLevel.Past); i < level.CandleIndex; i++ {
		if i < 0 {
			return false
		}
		if level.IsResistance() {
			if params.Candles[i].High > params.Candles[level.CandleIndex].High {
				return false
			}
		}

		if level.IsSupport() {
			if params.Candles[i].Low < params.Candles[level.CandleIndex].Low {
				return false
			}
		}
	}

	return true
}

func GetServiceInstance() Interface {
	return &Service{}
}
