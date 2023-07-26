package emaCrossover

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/types"
	"TradingBot/src/utils"
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
	PriceOffset                     float64
	CandlesAmountForHorizontalLevel *types.CandlesAmountForHorizontalLevel
	Candles                         []*types.Candle
	MaxAttempts                     int
	GetHorizontalLevel              func(horizontalLevels.GetLevelParams) (*horizontalLevels.Level, error)
	Log                             func(m string)
}

func getStopLoss(params GetStopLossParams) float64 {
	params.Log("getStopLoss called with this params -> " +
		"LongOrShort -> " + params.LongOrShort + " " +
		"PositionPrice -> " + utils.FloatToString(params.PositionPrice, 5) + " " +
		"MinStopLossDistance -> " + utils.FloatToString(params.MinStopLossDistance, 5) + " " +
		"MaxStopLossDistance -> " + utils.FloatToString(params.MaxStopLossDistance, 5) + " " +
		"PriceOffset -> " + utils.FloatToString(params.PriceOffset, 5) + " " +
		"CandlesAmountForHorizontalLevel -> " + utils.GetStringRepresentation(params.CandlesAmountForHorizontalLevel) + " ",
	)

	var sl float64
	var attempt int = 0
	var isValidSL bool = false
	index := int64(len(params.Candles) - 1)
	for {
		level, err := params.GetHorizontalLevel(horizontalLevels.GetLevelParams{
			StartAt: index,
			CandlesAmountToBeConsideredHorizontalLevel: *params.CandlesAmountForHorizontalLevel,
			Candles:        params.Candles,
			CandlesToCheck: 300,
		})
		if err != nil {
			params.Log(params.LongOrShort + " POSITION | Error when getting the horizontal level price for the SL -> " + err.Error())
			break
		}
		params.Log(params.LongOrShort + " POSITION | Horizontal level is -> " + utils.GetStringRepresentation(level))

		if level.Type == horizontalLevels.RESISTANCE_TYPE {
			sl = level.Candle.High
		}
		if level.Type == horizontalLevels.SUPPORT_TYPE {
			sl = level.Candle.Low
		}

		if params.LongOrShort == "long" {
			sl = sl + params.PriceOffset
		} else {
			sl = sl - params.PriceOffset
		}
		params.Log(params.LongOrShort + " POSITION | SL price after offset is -> " + utils.FloatToString(sl, 5))

		if params.LongOrShort == "long" {
			isValidSL = !((err != nil || sl >= params.PositionPrice) || (params.PositionPrice-sl >= params.MaxStopLossDistance) || (params.PositionPrice-sl <= params.MinStopLossDistance))
		} else {
			isValidSL = !((err != nil || sl <= params.PositionPrice) || (sl-params.PositionPrice >= params.MaxStopLossDistance) || (sl-params.PositionPrice <= params.MinStopLossDistance))
		}

		if isValidSL {
			break
		}

		params.Log(params.LongOrShort + " POSITION | Invalid SL, trying again ...")
		index = level.CandleIndex - 1
		attempt++

		if attempt == params.MaxAttempts {
			params.Log(params.LongOrShort + " POSITION | Max attempts reached - Unable to find a propert stop loss.")
			break
		}
	}

	if isValidSL {
		return sl
	}

	if params.LongOrShort == "long" {
		return params.PositionPrice - params.MaxStopLossDistance
	} else {
		return params.PositionPrice + params.MaxStopLossDistance
	}

}

func getEma(candle *types.Candle, candlesAmount int64) types.MovingAverage {
	return funk.Find(candle.Indicators.MovingAverages, func(ma types.MovingAverage) bool {
		return ma.CandlesAmount == candlesAmount
	}).(types.MovingAverage)
}
