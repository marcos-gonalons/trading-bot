package horizontalLevels

// Interface ...
type Interface interface {
	GetResistancePrice(candlesWithLowerPriceToBeConsideredTop int) (float64, error)
	GetSupportPrice(candlesWithHigherPriceToBeConsideredBottom int) (float64, error)
}
