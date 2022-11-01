package emaCrossover

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/types"
	"time"

	"github.com/thoas/go-funk"
)

const BASE_EMA = 200
const SMALL_EMA = 9
const BIG_EMA = 21

func closePositionOnReversal(
	position *api.Position,
	lastCandle *types.Candle,
	minProfit float64,
	API api.Interface,
	APIRetryFacade retryFacade.Interface,
	marketData *types.MarketData,
	log func(m string),
) {

	log("Checking if the position must be closed on trend reversal ...")
	if API.IsLongPosition(position) &&
		lastCandle.Close-position.AvgPrice > minProfit &&
		getEma(lastCandle, SMALL_EMA).Value < getEma(lastCandle, BIG_EMA).Value {

		log("Small EMA crossed below the big EMA, and the price is above the min profit. Closing the long position ...")
		APIRetryFacade.ClosePosition(position.Instrument, retryFacade.RetryParams{
			DelayBetweenRetries: 5 * time.Second,
			MaxRetries:          20,
		})
		API.AddTrade(
			nil,
			position,
			func(price float64, order *api.Order) float64 {
				return price
			},
			marketData.EurExchangeRate,
			lastCandle,
			marketData,
		)
		return
	}

	if API.IsShortPosition(position) &&
		position.AvgPrice-lastCandle.Close > minProfit &&
		getEma(lastCandle, SMALL_EMA).Value > getEma(lastCandle, BIG_EMA).Value {

		log("Small EMA crossed above the big EMA, and the price is above the min profit. Closing the short position ...")
		APIRetryFacade.ClosePosition(position.Instrument, retryFacade.RetryParams{
			DelayBetweenRetries: 5 * time.Second,
			MaxRetries:          20,
		})
		API.AddTrade(
			nil,
			position,
			func(price float64, order *api.Order) float64 {
				return price
			},
			marketData.EurExchangeRate,
			lastCandle,
			marketData,
		)
		return
	}
}

type GetStopLossParams struct {
	LongOrShort                     string
	PositionPrice                   float64
	MinStopLossDistance             float64
	MaxStopLossDistance             float64
	CandleIndex                     int
	PriceOffset                     float64
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
			sl = sl + params.PriceOffset
			if err != nil || sl >= params.PositionPrice {
				return params.PositionPrice - params.MaxStopLossDistance
			}
			if params.PositionPrice-sl >= params.MaxStopLossDistance {
				return params.PositionPrice - params.MaxStopLossDistance
			}
			if params.PositionPrice-sl <= params.MinStopLossDistance {
				return params.PositionPrice - params.MinStopLossDistance
			}
			return sl
		}
		if params.LongOrShort == "short" {
			sl, err := params.GetResistancePrice(
				*params.CandlesAmountForHorizontalLevel,
				params.CandleIndex,
				params.Candles,
			)
			sl = sl - params.PriceOffset
			if err != nil || sl <= params.PositionPrice {
				return params.PositionPrice + params.MaxStopLossDistance
			}
			if sl-params.PositionPrice >= params.MaxStopLossDistance {
				return params.PositionPrice + params.MaxStopLossDistance
			}
			if sl-params.PositionPrice <= params.MinStopLossDistance {
				return params.PositionPrice + params.MinStopLossDistance
			}
			return sl
		}
	} else {
		if params.LongOrShort == "long" {
			return params.PositionPrice - params.MaxStopLossDistance
		}
		if params.LongOrShort == "short" {
			return params.PositionPrice + params.MaxStopLossDistance
		}
	}

	return 0.
}

func getEma(candle *types.Candle, candlesAmount int64) types.MovingAverage {
	return funk.Find(candle.Indicators.MovingAverages, func(ma types.MovingAverage) bool {
		return ma.CandlesAmount == candlesAmount
	}).(types.MovingAverage)
}
