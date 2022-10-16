package strategies

import (
	"TradingBot/src/markets/baseMarketClass"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/utils"
	"time"
)

// todo: code is very similar between strategies, find a way to reuse code
// things like checking valid time, the pending orders logic and CheckIfSLShouldBeAdjusted/CheckOpenPositionTTL are the same for every strategy
// so maybe move that code somewhere else so it can be shared between all the strategies
// idea: parent strategy class with method "Execute"
// This method accepts a func() with the custom strategy code. Before calling this code, we will call the shared code

// ResistanceBounce ...
func ResistanceBounce(params StrategyParams) {
	var strategyName = params.BaseMarketClass.Name + " - RB"
	var log = func(msg string) {
		params.BaseMarketClass.Log(strategyName, msg)
	}

	log("resistanceBounce started")
	defer func() {
		log("resistanceBounce ended")
	}()

	validMonths := params.MarketStrategyParams.ValidTradingTimes.ValidMonths
	validWeekdays := params.MarketStrategyParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := params.MarketStrategyParams.ValidTradingTimes.ValidHalfHours

	if !utils.IsExecutionTimeValid(params.BaseMarketClass.GetCurrentExecutionTime(), validMonths, []string{}, []string{}) || !utils.IsExecutionTimeValid(params.BaseMarketClass.GetCurrentExecutionTime(), []string{}, validWeekdays, []string{}) {
		log("Today it's not the day for resistance bounce for " + params.BaseMarketClass.Market.SocketName)
		return
	}

	isValidTimeToOpenAPosition := utils.IsExecutionTimeValid(
		params.BaseMarketClass.GetCurrentExecutionTime(),
		validMonths,
		validWeekdays,
		validHalfHours,
	)

	if params.WithPendingOrders {
		if !isValidTimeToOpenAPosition {
			params.BaseMarketClass.SavePendingOrder(ibroker.ShortSide, params.MarketStrategyParams.ValidTradingTimes)
		} else {
			if params.BaseMarketClass.GetPendingOrder() != nil {
				params.BaseMarketClass.CreatePendingOrder(ibroker.ShortSide)
			}
			params.BaseMarketClass.SetPendingOrder(nil)
		}
	}

	p := utils.FindPositionByMarket(params.BaseMarketClass.APIData.GetPositions(), params.BaseMarketClass.GetMarket().BrokerAPIName)
	if p != nil && p.Side == ibroker.ShortSide {
		params.BaseMarketClass.CheckIfSLShouldBeAdjusted(params.MarketStrategyParams, p)
		params.BaseMarketClass.CheckOpenPositionTTL(params.MarketStrategyParams, p)
	}

	lastCompletedCandleIndex := len(params.BaseMarketClass.CandlesHandler.GetCandles()) - 2
	price, err := params.BaseMarketClass.HorizontalLevelsService.GetResistancePrice(*params.MarketStrategyParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good short setup yet -> " + err.Error()
		log(errorMessage)
		return
	}

	price = price - params.MarketStrategyParams.LimitAndStopOrderPriceOffset
	if price <= float64(params.BaseMarketClass.GetCurrentBrokerQuote().Ask) {
		log("Price is lower than the current ask, so we can't create the short order now. Price is -> " + utils.FloatToString(price, params.BaseMarketClass.GetMarket().PriceDecimals))
		log("Quote is -> " + utils.GetStringRepresentation(params.BaseMarketClass.GetCurrentBrokerQuote()))
		return
	}

	log("Ok, we might have a short setup at price " + utils.FloatToString(price, params.BaseMarketClass.GetMarket().PriceDecimals))
	if !params.BaseMarketClass.TrendsService.IsBearishTrend(
		params.MarketStrategyParams.TrendCandles,
		params.MarketStrategyParams.TrendDiff,
		params.BaseMarketClass.CandlesHandler.GetCandles(),
		lastCompletedCandleIndex,
	) {
		log("At the end it wasn't a good short setup, doing nothing ...")

		if params.CloseOrdersOnBadTrend && utils.FindPositionByMarket(params.BaseMarketClass.APIData.GetPositions(), params.BaseMarketClass.GetMarket().BrokerAPIName) == nil {
			log("There isn't an open position, closing short orders ...")
			params.BaseMarketClass.APIRetryFacade.CloseOrders(
				params.BaseMarketClass.API.GetWorkingOrderWithBracketOrders(ibroker.ShortSide, params.BaseMarketClass.GetMarket().BrokerAPIName, params.BaseMarketClass.APIData.GetOrders()),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
				},
			)
		}

		return
	}

	onValidTradeSetupParams := baseMarketClass.OnValidTradeSetupParams{
		Price:              price,
		StopLossDistance:   params.MarketStrategyParams.StopLossDistance,
		TakeProfitDistance: params.MarketStrategyParams.TakeProfitDistance,
		RiskPercentage:     params.MarketStrategyParams.RiskPercentage,
		IsValidTime:        isValidTimeToOpenAPosition,
		Side:               ibroker.ShortSide,
		StrategyName:       strategyName,
		WithPendingOrders:  params.WithPendingOrders,
		OrderType:          ibroker.LimitType,
		MinPositionSize:    params.MarketStrategyParams.MinPositionSize,
	}

	if utils.FindPositionByMarket(params.BaseMarketClass.APIData.GetPositions(), params.BaseMarketClass.GetMarket().BrokerAPIName) != nil {
		log("There is an open position, no need to close any orders ...")
		params.BaseMarketClass.OnValidTradeSetup(onValidTradeSetupParams)
	} else {
		log("There isn't any open position. Closing orders first ...")
		params.BaseMarketClass.APIRetryFacade.CloseOrders(
			params.BaseMarketClass.API.GetWorkingOrders(utils.FilterOrdersByMarket(params.BaseMarketClass.APIData.GetOrders(), params.BaseMarketClass.GetMarket().BrokerAPIName)),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					params.BaseMarketClass.OnValidTradeSetup(onValidTradeSetupParams)
				},
			},
		)
	}
}
