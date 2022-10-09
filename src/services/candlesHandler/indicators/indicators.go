package indicators

import (
	"TradingBot/src/services/candlesHandler/indicators/movingAverage"
	"TradingBot/src/types"
)

type Indicators struct {
	IndicatorsServices []IndicatorsInterface
}

func (s *Indicators) AddIndicators(candles []*types.Candle) {
	for _, i := range s.IndicatorsServices {
		i.AddData(candles)
	}
}

// GetInstance ...
func GetInstance() MainInterface {
	movingAverages := &movingAverage.MovingAverage{}

	return &Indicators{
		IndicatorsServices: []IndicatorsInterface{movingAverages},
	}
}
