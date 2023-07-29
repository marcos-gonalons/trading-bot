package ema

import (
	"TradingBot/src/types"

	"github.com/thoas/go-funk"
)

type Service struct{}

func (s *Service) IsPriceAboveEMA(candle *types.Candle, emaCandles int64) bool {
	return candle.Close >= s.GetEma(candle, emaCandles).Value
}

func (s *Service) IsPriceBelowEMA(candle *types.Candle, emaCandles int64) bool {
	return candle.Close <= s.GetEma(candle, emaCandles).Value
}

func (s *Service) GetEma(candle *types.Candle, candlesAmount int64) types.MovingAverage {
	return funk.Find(candle.Indicators.MovingAverages, func(ma types.MovingAverage) bool {
		return ma.CandlesAmount == candlesAmount
	}).(types.MovingAverage)
}

func GetServiceInstance() Interface {
	return &Service{}
}
