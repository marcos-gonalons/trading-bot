package candlesHandler

import (
	"TradingBot/src/types"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// Interface ...
type Interface interface {
	InitCandles()
	UpdateCandles(
		data *tradingviewsocket.QuoteData,
		currentExecutionTime time.Time,
		previousExecutionTime time.Time,
		lastVolume float64,
	)
	GetCandles() []*types.Candle
	GetLastCandle() *types.Candle
}
