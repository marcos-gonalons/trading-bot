package indicators

import "TradingBot/src/types"

type MainInterface interface {
	AddIndicators(candles []*types.Candle, lastCandleOnly bool)
}

type IndicatorsInterface interface {
	AddData(candles []*types.Candle, lastCandleOnly bool)
}
