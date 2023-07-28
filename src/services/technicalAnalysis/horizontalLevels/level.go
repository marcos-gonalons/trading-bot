package horizontalLevels

import "TradingBot/src/types"

type Level struct {
	Candle      *types.Candle
	CandleIndex int64
	Type        types.LevelType
}

func GetOpposite(t types.LevelType) types.LevelType {
	if t == types.RESISTANCE_TYPE {
		return types.SUPPORT_TYPE
	}
	if t == types.SUPPORT_TYPE {
		return types.RESISTANCE_TYPE
	}
	panic("Invalid level type")
}
func (l *Level) IsSupport() bool {
	return l.Type == types.SUPPORT_TYPE
}
func (l *Level) IsResistance() bool {
	return l.Type == types.RESISTANCE_TYPE
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
