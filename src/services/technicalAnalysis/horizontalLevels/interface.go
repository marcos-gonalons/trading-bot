package horizontalLevels

import "TradingBot/src/types"

// Interface ...
type Interface interface {
	GetResistancePrice(
		candlesWithLowerPriceToBeConsideredTop types.CandlesAmountForHorizontalLevel,
		lastCompletedCandleIndex int,
		candles []*types.Candle,
	) (float64, error)
	GetSupportPrice(
		candlesWithHigherPriceToBeConsideredBottom types.CandlesAmountForHorizontalLevel,
		lastCompletedCandleIndex int,
		candles []*types.Candle,
	) (float64, error)
}
