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

func (l *Level) IsSupport() bool {
	return l.Type == SUPPORT_TYPE
}
func (l *Level) IsResistance() bool {
	return l.Type == RESISTANCE_TYPE
}
func (l *Level) GetPrice() float64 {
	if l.IsResistance() {
		return l.Candle.High
	}
	if l.IsSupport() {
		return l.Candle.Low
	}
	panic("Invalid level type")
}
