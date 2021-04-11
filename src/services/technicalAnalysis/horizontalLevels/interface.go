package horizontalLevels

// Interface ...
type Interface interface {
	GetResistancePrice(
		candlesWithLowerPriceToBeConsideredTop int,
		lastCompletedCandleIndex int,
	) (float64, error)
	GetSupportPrice(
		candlesWithHigherPriceToBeConsideredBottom int,
		lastCompletedCandleIndex int,
	) (float64, error)
}
