package strategies

import (
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/strategies/tickers/baseTickerClass"
	"TradingBot/src/utils"
	"time"
)

// ResistanceBreakoutAnticipation ...
func ResistanceBreakoutAnticipation(params StrategyParams) {
	var strategyName = params.BaseTickerClass.Name + " - RBA"
	var log = func(msg string) {
		params.BaseTickerClass.Log(strategyName, msg)
	}

	log("resistanceBreakoutAnticipationStrategy started")
	defer func() {
		log("resistanceBreakoutAnticipationStrategy ended")
	}()

	validMonths := params.TickerStrategyParams.ValidTradingTimes.ValidMonths
	validWeekdays := params.TickerStrategyParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := params.TickerStrategyParams.ValidTradingTimes.ValidHalfHours

	if !utils.IsExecutionTimeValid(params.BaseTickerClass.GetCurrentExecutionTime(), validMonths, []string{}, []string{}) || !utils.IsExecutionTimeValid(params.BaseTickerClass.GetCurrentExecutionTime(), []string{}, validWeekdays, []string{}) {
		log("Today it's not the day for resistance breakout anticipation for " + params.BaseTickerClass.Symbol.SocketName)
		return
	}

	isValidTimeToOpenAPosition := utils.IsExecutionTimeValid(
		params.BaseTickerClass.GetCurrentExecutionTime(),
		validMonths,
		validWeekdays,
		validHalfHours,
	)

	if params.WithPendingOrders {
		if !isValidTimeToOpenAPosition {
			params.BaseTickerClass.SavePendingOrder(ibroker.LongSide, params.TickerStrategyParams.ValidTradingTimes)
		} else {
			if params.BaseTickerClass.GetPendingOrder() != nil {
				params.BaseTickerClass.CreatePendingOrder(ibroker.LongSide)
			}
			params.BaseTickerClass.SetPendingOrder(nil)
		}
	}

	p := utils.FindPositionBySymbol(params.BaseTickerClass.GetPositions(), params.BaseTickerClass.GetSymbol().BrokerAPIName)
	if p != nil && p.Side == ibroker.LongSide {
		params.BaseTickerClass.CheckIfSLShouldBeAdjusted(params.TickerStrategyParams, p)
	}

	lastCompletedCandleIndex := len(params.BaseTickerClass.CandlesHandler.GetCandles()) - 2
	price, err := params.BaseTickerClass.HorizontalLevelsService.GetResistancePrice(params.TickerStrategyParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good long setup yet -> " + err.Error()
		log(errorMessage)
		return
	}

	price = price - params.TickerStrategyParams.PriceOffset
	if price <= float64(params.BaseTickerClass.GetCurrentBrokerQuote().Ask) {
		log("Price is lower than the current ask, so we can't create the long order now. Price is -> " + utils.FloatToString(price, 2))
		log("Quote is -> " + utils.GetStringRepresentation(params.BaseTickerClass.GetCurrentBrokerQuote()))
		return
	}

	log("Ok, we might have a long setup at price " + utils.FloatToString(price, 2))
	if !params.BaseTickerClass.TrendsService.IsBullishTrend(
		params.TickerStrategyParams.TrendCandles,
		params.TickerStrategyParams.TrendDiff,
		params.BaseTickerClass.CandlesHandler.GetCandles(),
		lastCompletedCandleIndex,
	) {
		log("At the end it wasn't a good long setup, doing nothing ...")

		if params.CloseOrdersOnBadTrend && utils.FindPositionBySymbol(params.BaseTickerClass.GetPositions(), params.BaseTickerClass.GetSymbol().BrokerAPIName) == nil {
			log("There isn't an open position, closing long orders ...")
			params.BaseTickerClass.APIRetryFacade.CloseOrders(
				params.BaseTickerClass.API.GetWorkingOrderWithBracketOrders(ibroker.LongSide, params.BaseTickerClass.GetSymbol().BrokerAPIName, params.BaseTickerClass.GetOrders()),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback:     func() { params.BaseTickerClass.SetOrders(nil) },
				},
			)
		}

		return
	}

	onValidTradeSetupParams := baseTickerClass.OnValidTradeSetupParams{
		Price:              price,
		StopLossDistance:   params.TickerStrategyParams.StopLossDistance,
		TakeProfitDistance: params.TickerStrategyParams.TakeProfitDistance,
		RiskPercentage:     params.TickerStrategyParams.RiskPercentage,
		IsValidTime:        isValidTimeToOpenAPosition,
		Side:               ibroker.LongSide,
		StrategyName:       strategyName,
		WithPendingOrders:  params.WithPendingOrders,
		OrderType:          ibroker.StopType,
	}

	if utils.FindPositionBySymbol(params.BaseTickerClass.GetPositions(), params.BaseTickerClass.GetSymbol().BrokerAPIName) != nil {
		log("There is an open position, no need to close any orders ...")
		params.BaseTickerClass.OnValidTradeSetup(onValidTradeSetupParams)
	} else {
		log("There isn't any open position. Closing orders first ...")
		params.BaseTickerClass.APIRetryFacade.CloseOrders(
			params.BaseTickerClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(params.BaseTickerClass.GetOrders(), params.BaseTickerClass.GetSymbol().BrokerAPIName)),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					params.BaseTickerClass.OnValidTradeSetup(onValidTradeSetupParams)
				},
			},
		)
	}
}
