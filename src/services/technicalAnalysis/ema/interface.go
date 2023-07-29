package ema

import "TradingBot/src/types"

// Interface ...
type Interface interface {
	IsPriceAboveEMA(candle *types.Candle, emaCandles int64) bool
	IsPriceBelowEMA(candle *types.Candle, emaCandles int64) bool
	GetEma(candle *types.Candle, candlesAmount int64) types.MovingAverage
}
