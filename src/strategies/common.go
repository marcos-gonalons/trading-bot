package strategies

import (
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"time"
)

type HandleTrailingSLAndTPParams struct {
	TrailingSL *types.TrailingStopLoss
	TrailingTP *types.TrailingTakeProfit
	Position   *api.Position
	LastCandle *types.Candle
	MarketData *types.MarketData
	Container  *services.Container
	Log        func(msg string)
}

func HandleTrailingSLAndTP(params HandleTrailingSLAndTPParams) {
	handleTrailingSL := func() {
		params.Log("Checking if the position needs to have the SL adjusted with this params ... " + utils.GetStringRepresentation(params))
		params.Log("Position is " + utils.GetStringRepresentation(params.Position))

		_, tpOrder := params.Container.API.GetBracketOrdersForOpenedPosition(params.Position)

		if tpOrder == nil {
			params.Log("Take Profit order not found ...")
			return
		}

		shouldBeAdjusted := false
		var newSL float64
		if params.Container.API.IsLongPosition(params.Position) {
			shouldBeAdjusted = float64(*tpOrder.LimitPrice)-params.LastCandle.High < params.TrailingSL.TPDistanceShortForTighterSL
			newSL = float64(params.Position.AvgPrice) + params.TrailingSL.SLDistanceWhenTPIsVeryClose

			if shouldBeAdjusted && newSL >= params.LastCandle.Close {
				params.Log("Can't adjust the SL for the long position since the new SL is higher than the current price")
				return
			}
		} else {
			shouldBeAdjusted = params.LastCandle.Low-float64(*tpOrder.LimitPrice) < params.TrailingSL.TPDistanceShortForTighterSL
			newSL = float64(params.Position.AvgPrice) - params.TrailingSL.SLDistanceWhenTPIsVeryClose

			if shouldBeAdjusted && newSL <= params.LastCandle.Close {
				params.Log("Can't adjust the SL for the short position since the new SL is lower than the current price")
				return
			}
		}

		if !shouldBeAdjusted {
			params.Log("The price is not close to the TP yet. Doing nothing ...")
			return
		}

		params.Log("The price is very close to the TP. Adjusting SL...")
		params.Container.APIRetryFacade.ModifyPosition(
			params.MarketData.BrokerAPIName,
			utils.FloatToString(float64(*tpOrder.LimitPrice), params.MarketData.PriceDecimals),
			utils.FloatToString(newSL, params.MarketData.PriceDecimals),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			},
		)
	}

	handleTrailingTP := func() {
		params.Log("Checking if the position needs to have the TP adjusted with this params ... " + utils.GetStringRepresentation(params))
		params.Log("Position is " + utils.GetStringRepresentation(params.Position))

		slOrder, _ := params.Container.API.GetBracketOrdersForOpenedPosition(params.Position)

		if slOrder == nil {
			params.Log("Take Profit order not found ...")
			return
		}

		shouldBeAdjusted := false
		var newTP float64
		if params.Container.API.IsLongPosition(params.Position) {
			shouldBeAdjusted = params.LastCandle.Low-float64(*slOrder.LimitPrice) < params.TrailingTP.SLDistanceShortForTighterTP
			newTP = float64(params.Position.AvgPrice) - params.TrailingTP.TPDistanceWhenSLIsVeryClose
			if shouldBeAdjusted && newTP <= params.LastCandle.Close {
				params.Log("Can't adjust the TP for the long position since the new TP is lower than the current price")
				return
			}
		} else {
			shouldBeAdjusted = float64(*slOrder.LimitPrice)-params.LastCandle.High < params.TrailingTP.SLDistanceShortForTighterTP
			newTP = float64(params.Position.AvgPrice) + params.TrailingTP.TPDistanceWhenSLIsVeryClose

			if shouldBeAdjusted && newTP >= params.LastCandle.Close {
				params.Log("Can't adjust the TP for the short position since the new TP is higher than the current price")
				return
			}
		}

		if !shouldBeAdjusted {
			params.Log("The price is not close to the SL yet. Doing nothing ...")
			return
		}

		params.Log("The price is very close to the SL. Adjusting SL...")
		params.Container.APIRetryFacade.ModifyPosition(
			params.MarketData.BrokerAPIName,
			utils.FloatToString(newTP, params.MarketData.PriceDecimals),
			utils.FloatToString(float64(*slOrder.LimitPrice), params.MarketData.PriceDecimals),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			},
		)
	}

	if params.TrailingSL != nil && params.TrailingSL.TPDistanceShortForTighterSL > 0 {
		handleTrailingSL()
	}

	if params.TrailingTP != nil && params.TrailingTP.SLDistanceShortForTighterTP > 0 {
		handleTrailingTP()
	}

}
