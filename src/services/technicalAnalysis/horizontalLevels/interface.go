package horizontalLevels

import "TradingBot/src/types"

// Interface ...
type Interface interface {
	GetResistancePrice(
		candlesWithLowerPriceToBeConsideredTop types.CandlesAmountForHorizontalLevel,
		candles []*types.Candle,
	) (float64, error)
	GetSupportPrice(
		candlesWithHigherPriceToBeConsideredBottom types.CandlesAmountForHorizontalLevel,
		candles []*types.Candle,
	) (float64, error)
}
