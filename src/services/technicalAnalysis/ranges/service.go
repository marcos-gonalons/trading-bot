package ranges

import (
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/types"
)

type Service struct {
	horizontalLevelsService horizontalLevels.Interface
}

func (s *Service) GetRange(params GetRangeParams) []*horizontalLevels.Level {
	levelToGet := params.StrategyParams.Ranges.StartWith
	index := int64(len(params.Candles) - 1)
	candlesToCheck := params.StrategyParams.Ranges.CandlesToCheck
	r := []*horizontalLevels.Level{}
	previousPotentialLevels := []*horizontalLevels.Level{}

	for len(r) < params.StrategyParams.Ranges.RangePoints {
		level := &horizontalLevels.Level{}
		level, previousPotentialLevels, r = s.getPreviousValidRangeLevel(
			1,
			index,
			levelToGet,
			previousPotentialLevels,
			candlesToCheck,
			r,
			params.StrategyParams,
			params.Candles,
			params.LastCompletedCandle,
		)

		if level == nil {
			break
		}
		r = append(r, level)

		index = level.CandleIndex - params.StrategyParams.Ranges.MinCandlesBetweenRangePoints
		candlesToCheck = params.StrategyParams.Ranges.MaxCandlesBetweenRangePoints
		levelToGet = horizontalLevels.GetOpposite(levelToGet)
	}

	if len(r) < params.StrategyParams.Ranges.RangePoints {
		return nil
	}

	return r
}

func (s *Service) GetAverages(r []*horizontalLevels.Level) (resistancesAverage float64, supportsAverage float64) {
	resistancesAmount := float64(0)
	supportsAmount := float64(0)
	totalResistancesPrices := float64(0)
	totalSupportsPrices := float64(0)

	for _, level := range r {
		if level.IsResistance() {
			totalResistancesPrices += level.GetPrice()
			resistancesAmount++
		}
		if level.IsSupport() {
			totalSupportsPrices += level.GetPrice()
			supportsAmount++
		}
	}

	resistancesAverage = totalResistancesPrices / resistancesAmount
	supportsAverage = totalSupportsPrices / supportsAmount

	return
}

func (s *Service) getPreviousValidRangeLevel(
	attempt int,
	startAt int64,
	levelType types.LevelType,
	previousPotentialLevels []*horizontalLevels.Level,
	candlesAmount int64,
	r []*horizontalLevels.Level,
	strategyParams types.MarketStrategyParams,
	candles []*types.Candle,
	lastCompletedCandle *types.Candle,
) (*horizontalLevels.Level, []*horizontalLevels.Level, []*horizontalLevels.Level) {
	indexToUse := startAt
	potentialLevels := []*horizontalLevels.Level{}

	for i := 0; i < 10; i++ {
		potentialLevel := s.horizontalLevelsService.GetLevel(levelType, horizontalLevels.GetLevelParams{
			StartAt: indexToUse,
			CandlesAmountToBeConsideredHorizontalLevel: strategyParams.CandlesAmountForHorizontalLevel,
			Candles:        candles,
			CandlesToCheck: strategyParams.Ranges.MaxCandlesBetweenRangePoints,
		})
		if potentialLevel == nil {
			continue
		}

		potentialLevels = append(potentialLevels, potentialLevel)

		indexToUse = potentialLevel.CandleIndex - 1
	}

	for _, potentialLevel := range potentialLevels {
		if !s.validateRangeLevel(potentialLevel, r, candles, lastCompletedCandle, &strategyParams.Ranges) {
			continue
		}

		return potentialLevel, potentialLevels, r
	}

	if len(previousPotentialLevels) == 0 || attempt == len(previousPotentialLevels) {
		return nil, nil, r
	}
	previousPotentialLevel := previousPotentialLevels[attempt]

	r = r[:len(r)-1]
	r = append(r, previousPotentialLevel)

	return s.getPreviousValidRangeLevel(
		attempt+1,
		previousPotentialLevel.CandleIndex-1,
		levelType,
		previousPotentialLevels,
		candlesAmount,
		r,
		strategyParams,
		candles,
		lastCompletedCandle,
	)
}

func (s *Service) validateRangeLevel(
	level *horizontalLevels.Level,
	r []*horizontalLevels.Level,
	candles []*types.Candle,
	lastCompletedCandle *types.Candle,
	validationParams *types.Ranges,
) bool {
	if len(r) == 0 {
		return true
	}

	toValidate := r
	toValidate = append(toValidate, level)

	return IsRangeValid(&IsRangeValidParams{
		Range:               toValidate,
		ValidationParams:    validationParams,
		Candles:             candles,
		LastCompletedCandle: lastCompletedCandle,
	})
}

func GetServiceInstance(horizontalLevelsService horizontalLevels.Interface) Interface {
	return &Service{
		horizontalLevelsService,
	}
}
