package emaCrossover

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
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
		// todo: API.AddTrade is only used for the simulator API, and it's always called after calling ClosePosition
		// refactor it: add the trade on ClosePosition method for the simulator API and remove AddTrade from the API interface.
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
	GetResistancePrice              func(types.CandlesAmountForHorizontalLevel, []*types.Candle) (float64, error)
	GetSupportPrice                 func(types.CandlesAmountForHorizontalLevel, []*types.Candle) (float64, error)
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
	if params.CandlesAmountForHorizontalLevel != nil {
		if params.LongOrShort == "long" {
			sl, err := params.GetSupportPrice(
				*params.CandlesAmountForHorizontalLevel,
				params.Candles,
			)
			if err != nil {
				params.Log("LONG POSITION | Error when getting the support price for the SL -> " + err.Error())
			}
			params.Log("LONG POSITION | Support price is -> " + utils.FloatToString(sl, 5))
			sl = sl + params.PriceOffset
			params.Log("LONG POSITION | SL price after offset is -> " + utils.FloatToString(sl, 5))
			if err != nil || sl >= params.PositionPrice {
				params.Log("LONG POSITION | Using MaxStopLossDistance because there was an error or the SL price is above the current position price")
				// todo: check using MINStopLossDistance here
				return params.PositionPrice - params.MaxStopLossDistance
			}
			if params.PositionPrice-sl >= params.MaxStopLossDistance {
				params.Log("LONG POSITION | Using MaxStopLossDistance because the SL distance is bigger than the MaxStopLossDistance")
				return params.PositionPrice - params.MaxStopLossDistance
			}
			if params.PositionPrice-sl <= params.MinStopLossDistance {
				params.Log("LONG POSITION | Using MinStopLossDistance because the SL distance is lower than the MinStopLossDistance")
				return params.PositionPrice - params.MinStopLossDistance
			}
			return sl
		}
		if params.LongOrShort == "short" {
			sl, err := params.GetResistancePrice(
				*params.CandlesAmountForHorizontalLevel,
				params.Candles,
			)
			if err != nil {
				params.Log("SHORT POSITION | Error when getting the resistance price for the SL -> " + err.Error())
			}
			params.Log("SHORT POSITION | Resistance price is -> " + utils.FloatToString(sl, 5))
			sl = sl - params.PriceOffset
			params.Log("SHORT POSITION | SL price after offset is -> " + utils.FloatToString(sl, 5))
			if err != nil || sl <= params.PositionPrice {
				params.Log("SHORT POSITION | Using MaxStopLossDistance because there was an error or the SL price is lower the current position price")
				// todo: check using MINStopLossDistance here
				return params.PositionPrice + params.MaxStopLossDistance
			}
			if sl-params.PositionPrice >= params.MaxStopLossDistance {
				params.Log("SHORT POSITION | Using MaxStopLossDistance because the SL distance is bigger than the MaxStopLossDistance")
				return params.PositionPrice + params.MaxStopLossDistance
			}
			if sl-params.PositionPrice <= params.MinStopLossDistance {
				params.Log("SHORT POSITION | Using MinStopLossDistance because the SL distance is lower than the MinStopLossDistance")
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
