package horizontalLevels

import "TradingBot/src/types"

// Interface ...
type Interface interface {
	GetLevel(levelType types.LevelType, params GetLevelParams) *Level
	GetResistance(params GetLevelParams) *Level
	GetSupport(params GetLevelParams) *Level
}

type GetLevelParams struct {
	StartAt                                    int64
	CandlesAmountToBeConsideredHorizontalLevel *types.CandlesAmountForHorizontalLevel
	Candles                                    []*types.Candle
	CandlesToCheck                             int64
}
