package horizontalLevels

import "TradingBot/src/types"

// Interface ...
type Interface interface {
	GetResistancePrice(
		candlesWithLowerPriceToBeConsideredTop types.CandlesAmountForHorizontalLevel,
		candles []*types.Candle,
		startingIndex int,
	) (float64, int, error)
	GetSupportPrice(
		candlesWithHigherPriceToBeConsideredBottom types.CandlesAmountForHorizontalLevel,
		candles []*types.Candle,
		startingIndex int,
	) (float64, int, error)
}
