package technicalanalysis

import (
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/types"
)

type Level struct {
	Candle types.Candle
	Index  int64
	Type   horizontalLevels.LevelType
}
