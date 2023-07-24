package ranges

import (
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/types"
)

// Interface ...
type Interface interface {
	GetRange(params GetRangeParams) ([]*horizontalLevels.Level, error)
	GetSupport(params GetRangeParams) ([]*horizontalLevels.Level, error)
}

type GetRangeParams struct {
	Candles          []*types.Candle
	CurrentCandle    *types.Candle
	CurrentDataIndex int64
	StrategyParams   types.MarketStrategyParams
}
