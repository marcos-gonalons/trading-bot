package horizontalLevels

import "TradingBot/src/types"

type LevelType string

const (
	RESISTANCE_TYPE LevelType = "resistance"
	SUPPORT_TYPE    LevelType = "support"
)

type Level struct {
	Candle      *types.Candle
	CandleIndex int64
	Type        LevelType
}
