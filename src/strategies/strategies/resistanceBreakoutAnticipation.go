package strategies

import (
	"TradingBot/src/markets/interfaces"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/utils"
	"time"
)

// ResistanceBreakoutAnticipation ...
func ResistanceBreakoutAnticipation(params StrategyParams) {
	var strategyName = params.MarketData.SocketName + " - RBA"
	var log = func(msg string) {
		params.Market.Log(strategyName, msg)
	}

	log("resistanceBreakoutAnticipation started")
	defer func() {
		log("resistanceBreakoutAnticipation ended")
	}()

	validMonths := params.MarketStrategyParams.ValidTradingTimes.ValidMonths
	validWeekdays := params.MarketStrategyParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := params.MarketStrategyParams.ValidTradingTimes.ValidHalfHours

	now := time.Now()
	if !utils.IsExecutionTimeValid(now, validMonths, []string{}, []string{}) || !utils.IsExecutionTimeValid(now, []string{}, validWeekdays, []string{}) {
		log("Today it's not the day for resistance breakout anticipation for " + params.MarketData.SocketName)
		return
	}

	isValidTimeToOpenAPosition := utils.IsExecutionTimeValid(
		now,
		validMonths,
		validWeekdays,
		validHalfHours,
	)

	if params.MarketStrategyParams.WithPendingOrders {
		if !isValidTimeToOpenAPosition {
			params.Market.SavePendingOrder(ibroker.LongSide, params.MarketStrategyParams.ValidTradingTimes)
		} else {
			if params.Market.GetPendingOrder() != nil {
				params.Market.CreatePendingOrder(ibroker.LongSide)
			}
			params.Market.SetPendingOrder(nil)
		}
	}

	p := utils.FindPositionByMarket(params.APIData.GetPositions(), params.MarketData.BrokerAPIName)
	if p != nil && p.Side == ibroker.LongSide {
		params.Market.CheckIfSLShouldBeAdjusted(params.MarketStrategyParams, p)
		params.Market.CheckOpenPositionTTL(params.MarketStrategyParams, p)
	}

	lastCompletedCandleIndex := len(params.CandlesHandler.GetCandles()) - 2
	price, err := params.HorizontalLevelsService.GetResistancePrice(*params.MarketStrategyParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good long setup yet -> " + err.Error()
		log(errorMessage)
		return
	}

	price = price - params.MarketStrategyParams.LimitAndStopOrderPriceOffset
	if price <= float64(params.Market.GetCurrentBrokerQuote().Ask) {
		log("Price is lower than the current ask, so we can't create the long order now. Price is -> " + utils.FloatToString(price, params.MarketData.PriceDecimals))
		log("Quote is -> " + utils.GetStringRepresentation(params.Market.GetCurrentBrokerQuote()))
		return
	}

	log("Ok, we might have a long setup at price " + utils.FloatToString(price, params.MarketData.PriceDecimals))
	if !params.TrendsService.IsBullishTrend(
		params.MarketStrategyParams.TrendCandles,
		params.MarketStrategyParams.TrendDiff,
		params.CandlesHandler.GetCandles(),
		lastCompletedCandleIndex,
	) {
		log("At the end it wasn't a good long setup, doing nothing ...")

		if params.MarketStrategyParams.CloseOrdersOnBadTrend && utils.FindPositionByMarket(params.APIData.GetPositions(), params.MarketData.BrokerAPIName) == nil {
			log("There isn't an open position, closing long orders ...")
			params.APIRetryFacade.CloseOrders(
				params.API.GetWorkingOrderWithBracketOrders(ibroker.LongSide, params.MarketData.BrokerAPIName, params.APIData.GetOrders()),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
				},
			)
		}

		return
	}

	onValidTradeSetupParams := interfaces.OnValidTradeSetupParams{
		Price:              price,
		StopLossDistance:   params.MarketStrategyParams.StopLossDistance,
		TakeProfitDistance: params.MarketStrategyParams.TakeProfitDistance,
		RiskPercentage:     params.MarketStrategyParams.RiskPercentage,
		IsValidTime:        isValidTimeToOpenAPosition,
		Side:               ibroker.LongSide,
		StrategyName:       strategyName,
		WithPendingOrders:  params.MarketStrategyParams.WithPendingOrders,
		OrderType:          ibroker.StopType,
		MinPositionSize:    params.MarketStrategyParams.MinPositionSize,
	}

	if utils.FindPositionByMarket(params.APIData.GetPositions(), params.MarketData.BrokerAPIName) != nil {
		log("There is an open position, no need to close any orders ...")
		params.Market.OnValidTradeSetup(onValidTradeSetupParams)
	} else {
		log("There isn't any open position. Closing orders first ...")
		params.APIRetryFacade.CloseOrders(
			params.API.GetWorkingOrders(utils.FilterOrdersByMarket(params.APIData.GetOrders(), params.MarketData.BrokerAPIName)),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					params.Market.OnValidTradeSetup(onValidTradeSetupParams)
				},
			},
		)
	}
}
