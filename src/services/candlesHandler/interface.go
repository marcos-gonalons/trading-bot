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
		data *tradingviewsocket.QuoteData,
		currentExecutionTime time.Time,
		lastVolume float64,
	)
	AddNewCandle(types.Candle)
	RemoveOldCandles(amount uint)
	GetCandles() []*types.Candle
	GetLastCandle() *types.Candle
}
