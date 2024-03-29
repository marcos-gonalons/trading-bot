package trends

import "TradingBot/src/types"

// Interface ...
type Interface interface {
	IsBullishTrend(
		candlesAmountToCheck int,
		priceGap float64,
		candles []*types.Candle,
	) bool
	IsBearishTrend(
		candlesAmountToCheck int,
		priceGap float64,
		candles []*types.Candle,
	) bool
}
