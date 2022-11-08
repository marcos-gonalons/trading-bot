package candlesHandler

import (
	"TradingBot/src/types"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// Interface ...
type Interface interface {
	InitCandles(currentExecutionTime time.Time, fileName string)
	UpdateCandles(data *tradingviewsocket.QuoteData, lastVolume float64, onNewCandleCallback func())
	AddNewCandle(types.Candle)
	RemoveOldCandles(amount uint)
	GetCompletedCandles() []*types.Candle
	GetLastCompletedCandle() *types.Candle
	SetCandles([]*types.Candle)
}
