package ranges

import (
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/types"
)

type GetRangeParams struct {
	Candles          []*types.Candle
	CurrentCandle    *types.Candle
	CurrentDataIndex int64
	StrategyParams   types.MarketStrategyParams
}

// Interface ...
type Interface interface {
	GetRange(params GetRangeParams) []*horizontalLevels.Level
	GetAverages(r []*horizontalLevels.Level) (resistancesAverage float64, supportsAverage float64)
}
