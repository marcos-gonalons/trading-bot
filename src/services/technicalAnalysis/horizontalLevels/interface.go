package horizontalLevels

import "TradingBot/src/types"

// Interface ...
type Interface interface {
	GetResistance(params GetLevelParams) (*Level, error)
	GetSupport(params GetLevelParams) (*Level, error)
}

type GetLevelParams struct {
	StartAt                                    int64
	CandlesAmountToBeConsideredHorizontalLevel types.CandlesAmountForHorizontalLevel
	Candles                                    []*types.Candle
	CandlesToCheck                             int64
}
