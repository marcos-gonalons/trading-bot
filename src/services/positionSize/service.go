package positionSize

import "math"

type Service struct {
	strategy uint
}

const BASED_ON_STOP_LOSS_DISTANCE = 0
const BASED_ON_MULTIPLIER = 1

func (s *Service) GetPositionSize(p GetPositionSizeParams) float32 {
	switch s.strategy {
	case BASED_ON_STOP_LOSS_DISTANCE:
		size := math.Floor((p.CurrentBalance*(p.RiskPercentage/100))/(p.StopLossDistance*p.MinPositionSize*p.EurExchangeRate)) * p.MinPositionSize
		if size == 0 {
			size = p.MinPositionSize
		}
		return float32(size)
	case BASED_ON_MULTIPLIER:
		var equityThresold = float64(3000)

		size := float32(math.Ceil(p.MinPositionSize * math.Floor(p.CurrentBalance/equityThresold) * float64(p.Multiplier)))
		if size == 0 {
			return float32(p.MinPositionSize)
		}
		return size
	default:
		panic("Invalid position size strategy")
	}

}

func (s *Service) SetStrategy(st uint) {
	s.strategy = st
}

func GetInstance() Interface {
	return &Service{
		strategy: BASED_ON_MULTIPLIER,
	}
}
