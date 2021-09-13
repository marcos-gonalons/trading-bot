package candlesHandler

import (
	"TradingBot/src/types"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// Interface ...
type Interface interface {
	InitCandles(currentExecutionTime time.Time, fileName string)
	UpdateCandles(
		symbol *types.Symbol,
		data *tradingviewsocket.QuoteData,
		currentExecutionTime time.Time,
		lastVolume float64,
	)
	RemoveOldCandles(amount uint)
	GetCandles() []*types.Candle
	GetLastCandle() *types.Candle
}
