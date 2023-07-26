package strategies

import (
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"errors"
	"time"
)

func OnBegin(params Params) (err error) {
	params.Market.Log("OnBegin... " + params.Type)
	var validMonths, validWeekdays, validHalfHours []string

	if params.MarketStrategyParams.ValidTradingTimes != nil {
		validMonths = params.MarketStrategyParams.ValidTradingTimes.ValidMonths
		validWeekdays = params.MarketStrategyParams.ValidTradingTimes.ValidWeekdays
		validHalfHours = params.MarketStrategyParams.ValidTradingTimes.ValidHalfHours
	}

	now := time.Now()
	if !utils.IsExecutionTimeValid(now, validMonths, []string{}, []string{}) || !utils.IsExecutionTimeValid(now, []string{}, validWeekdays, []string{}) {
		err = errors.New("not a valid month nor weekday to execute this strategy")
		return
	}

	isValidTimeToOpenAPosition := utils.IsExecutionTimeValid(
		now,
		validMonths,
		validWeekdays,
		validHalfHours,
	)

	if params.MarketStrategyParams.WithPendingOrders {
		params.Market.Log("Strategy has pending orders enabled.")
		if !isValidTimeToOpenAPosition {
			params.Market.Log("Not a valid time to open any position now. Calling SavePendingOrder ...")
			params.Market.SavePendingOrder(params.Type, params.MarketStrategyParams.ValidTradingTimes)
		} else {
			params.Market.Log("Time is valid")
			if params.Market.GetPendingOrder() != nil {
				params.Market.Log("There is a pending order. Creating the pending order ...")
				params.Market.CreatePendingOrder(params.Type)
			}
			params.Market.SetPendingOrder(nil)
		}
	}

	p := utils.FindPositionByMarket(services.GetServicesContainer().APIData.GetPositions(), params.MarketData.BrokerAPIName)
	if p != nil && p.Side == params.Type {
		params.Market.Log("There is an open position. Calling HandleTrailingSLAndTP and CheckOpenPositionTTL ...")
		params.Market.Log("Position is -> " + utils.GetStringRepresentation(p))
		HandleTrailingSLAndTP(HandleTrailingSLAndTPParams{
			TrailingSL: params.MarketStrategyParams.TrailingStopLoss,
			TrailingTP: params.MarketStrategyParams.TrailingTakeProfit,
			Position:   p,
			LastCandle: params.CandlesHandler.GetLastCompletedCandle(),
			MarketData: params.MarketData,
			Container:  services.GetServicesContainer(),
		}, params.Market.Log)
		params.Market.CheckOpenPositionTTL(params.MarketStrategyParams, p)
	}

	return
}

type HandleTrailingSLAndTPParams struct {
	TrailingSL *types.TrailingStopLoss
	TrailingTP *types.TrailingTakeProfit
	Position   *api.Position
	LastCandle *types.Candle
	MarketData *types.MarketData
	Container  *services.Container
}

func HandleTrailingSLAndTP(
	params HandleTrailingSLAndTPParams,
	log func(msg string),
) {
	handleTrailingSL := func() {
		log("Checking if the position needs to have the SL adjusted with this params ... " + utils.GetStringRepresentation(params))

		_, tpOrder := services.GetServicesContainer().API.GetBracketOrders(params.Position.Instrument)

		if tpOrder == nil {
			log("Take Profit order not found ...")
			return
		}

		shouldBeAdjusted := false
		var newSL float64
		if services.GetServicesContainer().API.IsLongPosition(params.Position) {
			shouldBeAdjusted = *tpOrder.LimitPrice-params.LastCandle.High < params.TrailingSL.TPDistanceShortForTighterSL
			newSL = params.Position.AvgPrice + params.TrailingSL.SLDistanceWhenTPIsVeryClose

			if shouldBeAdjusted && newSL >= params.LastCandle.Close {
				log("Can't adjust the SL for the long position since the new SL is higher than the current price")
				log("New SL that will not be applied -> " + utils.FloatToString(newSL, params.MarketData.PriceDecimals))
				return
			}
		} else {
			shouldBeAdjusted = params.LastCandle.Low-*tpOrder.LimitPrice < params.TrailingSL.TPDistanceShortForTighterSL
			newSL = params.Position.AvgPrice - params.TrailingSL.SLDistanceWhenTPIsVeryClose

			if shouldBeAdjusted && newSL <= params.LastCandle.Close {
				log("Can't adjust the SL for the short position since the new SL is lower than the current price")
				log("New SL that will not be applied -> " + utils.FloatToString(newSL, params.MarketData.PriceDecimals))
				return
			}
		}

		if !shouldBeAdjusted {
			log("The price is not close to the TP yet. Doing nothing ...")
			return
		}

		log("The price is very close to the TP. Adjusting SL...")
		log("New SL -> " + utils.FloatToString(newSL, params.MarketData.PriceDecimals))
		services.GetServicesContainer().APIRetryFacade.ModifyPosition(
			params.MarketData.BrokerAPIName,
			utils.FloatToString(*tpOrder.LimitPrice, params.MarketData.PriceDecimals),
			utils.FloatToString(newSL, params.MarketData.PriceDecimals),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			},
		)
	}

	handleTrailingTP := func() {
		log("Checking if the position needs to have the TP adjusted with this params ... " + utils.GetStringRepresentation(params))

		slOrder, _ := services.GetServicesContainer().API.GetBracketOrders(params.Position.Instrument)

		if slOrder == nil {
			log("Stop Loss order not found ...")
			return
		}

		shouldBeAdjusted := false
		var newTP float64
		if services.GetServicesContainer().API.IsLongPosition(params.Position) {
			shouldBeAdjusted = params.LastCandle.Low-*slOrder.StopPrice < params.TrailingTP.SLDistanceShortForTighterTP
			newTP = params.Position.AvgPrice - params.TrailingTP.TPDistanceWhenSLIsVeryClose
			if shouldBeAdjusted && newTP <= params.LastCandle.Close {
				log("Can't adjust the TP for the long position since the new TP is lower than the current price")
				log("New TP that will not be applied -> " + utils.FloatToString(newTP, params.MarketData.PriceDecimals))
				return
			}
		} else {
			shouldBeAdjusted = *slOrder.StopPrice-params.LastCandle.High < params.TrailingTP.SLDistanceShortForTighterTP
			newTP = params.Position.AvgPrice + params.TrailingTP.TPDistanceWhenSLIsVeryClose

			if shouldBeAdjusted && newTP >= params.LastCandle.Close {
				log("Can't adjust the TP for the short position since the new TP is higher than the current price")
				log("New TP that will not be applied -> " + utils.FloatToString(newTP, params.MarketData.PriceDecimals))
				return
			}
		}

		if !shouldBeAdjusted {
			log("The price is not close to the SL yet. Doing nothing ...")
			return
		}

		log("The price is very close to the SL. Adjusting TP...")
		log("New TP -> " + utils.FloatToString(newTP, params.MarketData.PriceDecimals))
		services.GetServicesContainer().APIRetryFacade.ModifyPosition(
			params.MarketData.BrokerAPIName,
			utils.FloatToString(newTP, params.MarketData.PriceDecimals),
			utils.FloatToString(*slOrder.StopPrice, params.MarketData.PriceDecimals),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			},
		)
	}

	log("Position is " + utils.GetStringRepresentation(params.Position))
	log("LastCandle is " + utils.GetStringRepresentation(params.LastCandle))

	if params.TrailingSL != nil && params.TrailingSL.TPDistanceShortForTighterSL > 0 {
		handleTrailingSL()
	}

	if params.TrailingTP != nil && params.TrailingTP.SLDistanceShortForTighterTP > 0 {
		handleTrailingTP()
	}

}
