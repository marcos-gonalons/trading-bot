package technicalanalysis

import (
	"TradingBot/src/types"
)

type Level struct {
	Candle types.Candle
	Index  int64
	Type   types.LevelType
}
