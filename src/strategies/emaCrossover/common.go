package emaCrossover

import (
	"TradingBot/src/services/api"
	"TradingBot/src/types"

	"github.com/thoas/go-funk"
)

const BASE_EMA = 200
const SMALL_EMA = 9
const BIG_EMA = 21

func closePositionOnReversal(
	position *api.Position,
	lastCandle *types.Candle,
	minProfit float32,
	API api.Interface,
) {
	if API.IsLongPosition(position) &&
		lastCandle.Close-float64(position.AvgPrice) > float64(minProfit) &&
		getEma(lastCandle, SMALL_EMA).Value < getEma(lastCandle, BIG_EMA).Value {

		API.ClosePosition(position.Instrument)
		return
	}

	if API.IsShortPosition(position) &&
		float64(position.AvgPrice)-lastCandle.Close > float64(minProfit) &&
		getEma(lastCandle, SMALL_EMA).Value > getEma(lastCandle, BIG_EMA).Value {

		API.ClosePosition(position.Instrument)
		return
	}
}

type GetStopLossParams struct {
	LongOrShort                     string
	PositionPrice                   float32
	MinStopLossDistance             float32
	MaxStopLossDistance             float32
	CandleIndex                     int
	PriceOffset                     float32
	CandlesAmountForHorizontalLevel *types.CandlesAmountForHorizontalLevel
	Candles                         []*types.Candle
	GetResistancePrice              func(types.CandlesAmountForHorizontalLevel, int, []*types.Candle) (float64, error)
	GetSupportPrice                 func(types.CandlesAmountForHorizontalLevel, int, []*types.Candle) (float64, error)
}

func getStopLoss(params GetStopLossParams) float64 {
	if params.CandlesAmountForHorizontalLevel != nil {
		if params.LongOrShort == "long" {
			sl, err := params.GetSupportPrice(
				*params.CandlesAmountForHorizontalLevel,
				params.CandleIndex,
				params.Candles,
			)
			if err != nil || sl >= float64(params.PositionPrice) {
				return float64(params.PositionPrice) - float64(params.MaxStopLossDistance)
			}
			if float64(params.PositionPrice)-sl >= float64(params.MaxStopLossDistance) {
				return float64(params.PositionPrice) - float64(params.MaxStopLossDistance)
			}
			if float64(params.PositionPrice)-sl <= float64(params.MinStopLossDistance) {
				return float64(params.PositionPrice) - float64(params.MinStopLossDistance)
			}
			return sl + float64(params.PriceOffset)
		}
		if params.LongOrShort == "short" {
			sl, err := params.GetResistancePrice(
				*params.CandlesAmountForHorizontalLevel,
				params.CandleIndex,
				params.Candles,
			)
			if err != nil || sl <= float64(params.PositionPrice) {
				return float64(params.PositionPrice) + float64(params.MaxStopLossDistance)
			}
			if sl-float64(params.PositionPrice) >= float64(params.MaxStopLossDistance) {
				return float64(params.PositionPrice) + float64(params.MaxStopLossDistance)
			}
			if sl-float64(params.PositionPrice) <= float64(params.MinStopLossDistance) {
				return float64(params.PositionPrice) + float64(params.MinStopLossDistance)
			}
			return sl - float64(params.PriceOffset)
		}
	} else {
		if params.LongOrShort == "long" {
			return float64(params.PositionPrice) - float64(params.MaxStopLossDistance)
		}
		if params.LongOrShort == "short" {
			return float64(params.PositionPrice) + float64(params.MaxStopLossDistance)
		}
	}

	return 0.
}

func getEma(candle *types.Candle, candlesAmount int64) *types.MovingAverage {
	return funk.Find(candle.Indicators.MovingAverages, func(ma types.MovingAverage) bool {
		return ma.CandlesAmount == candlesAmount
	}).(*types.MovingAverage)
}
