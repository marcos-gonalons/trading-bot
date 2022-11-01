package positionSize

type GetPositionSizeParams struct {
	CurrentBalance   float64
	RiskPercentage   float64
	StopLossDistance float64
	MinPositionSize  float64
	EurExchangeRate  float64
	Multiplier       float64
}
type Interface interface {
	GetPositionSize(GetPositionSizeParams) float64
	SetStrategy(uint)
}
