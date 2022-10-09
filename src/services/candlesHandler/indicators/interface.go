package indicators

import "TradingBot/src/types"

type MainInterface interface {
	AddIndicators(candles []*types.Candle)
}

type IndicatorsInterface interface {
	AddData(candles []*types.Candle)
}
