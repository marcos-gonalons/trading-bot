package strategies

import (
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/strategies/markets/baseMarketClass"
	"TradingBot/src/utils"
	"time"
)

// SupportBounce ...
func SupportBounce(params StrategyParams) {
	var strategyName = params.BaseMarketClass.Name + " - SB"
	var log = func(msg string) {
		params.BaseMarketClass.Log(strategyName, msg)
	}

	log("supportBounce started")
	defer func() {
		log("supportBounce ended")
	}()

	validMonths := params.MarketStrategyParams.ValidTradingTimes.ValidMonths
	validWeekdays := params.MarketStrategyParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := params.MarketStrategyParams.ValidTradingTimes.ValidHalfHours

	if !utils.IsExecutionTimeValid(params.BaseMarketClass.GetCurrentExecutionTime(), validMonths, []string{}, []string{}) || !utils.IsExecutionTimeValid(params.BaseMarketClass.GetCurrentExecutionTime(), []string{}, validWeekdays, []string{}) {
		log("Today it's not the day for support bounce  for " + params.BaseMarketClass.Symbol.SocketName)
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
			params.BaseMarketClass.SavePendingOrder(ibroker.LongSide, params.MarketStrategyParams.ValidTradingTimes)
		} else {
			if params.BaseMarketClass.GetPendingOrder() != nil {
				params.BaseMarketClass.CreatePendingOrder(ibroker.LongSide)
			}
			params.BaseMarketClass.SetPendingOrder(nil)
		}
	}

	p := utils.FindPositionBySymbol(params.BaseMarketClass.APIData.GetPositions(), params.BaseMarketClass.GetSymbol().BrokerAPIName)
	if p != nil && p.Side == ibroker.LongSide {
		params.BaseMarketClass.CheckIfSLShouldBeAdjusted(params.MarketStrategyParams, p)
		params.BaseMarketClass.CheckOpenPositionTTL(params.MarketStrategyParams, p)
	}

	lastCompletedCandleIndex := len(params.BaseMarketClass.CandlesHandler.GetCandles()) - 2
	price, err := params.BaseMarketClass.HorizontalLevelsService.GetSupportPrice(*params.MarketStrategyParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good long setup yet -> " + err.Error()
		log(errorMessage)
		return
	}

	price = price + params.MarketStrategyParams.LimitAndStopOrderPriceOffset
	if price >= float64(params.BaseMarketClass.GetCurrentBrokerQuote().Bid) {
		log("Price is lower than the current ask, so we can't create the long order now. Price is -> " + utils.FloatToString(price, params.BaseMarketClass.GetSymbol().PriceDecimals))
		log("Quote is -> " + utils.GetStringRepresentation(params.BaseMarketClass.GetCurrentBrokerQuote()))
		return
	}

	log("Ok, we might have a long setup at price " + utils.FloatToString(price, params.BaseMarketClass.GetSymbol().PriceDecimals))
	if !params.BaseMarketClass.TrendsService.IsBullishTrend(
		params.MarketStrategyParams.TrendCandles,
		params.MarketStrategyParams.TrendDiff,
		params.BaseMarketClass.CandlesHandler.GetCandles(),
		lastCompletedCandleIndex,
	) {
		log("At the end it wasn't a good long setup")
		if params.CloseOrdersOnBadTrend && utils.FindPositionBySymbol(params.BaseMarketClass.APIData.GetPositions(), params.BaseMarketClass.GetSymbol().BrokerAPIName) == nil {
			log("There isn't an open position, closing long orders ...")
			params.BaseMarketClass.APIRetryFacade.CloseOrders(
				params.BaseMarketClass.API.GetWorkingOrderWithBracketOrders(ibroker.LongSide, params.BaseMarketClass.GetSymbol().BrokerAPIName, params.BaseMarketClass.APIData.GetOrders()),
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
		Side:               ibroker.LongSide,
		StrategyName:       strategyName,
		WithPendingOrders:  params.WithPendingOrders,
		OrderType:          ibroker.LimitType,
		MinPositionSize:    params.MarketStrategyParams.MinPositionSize,
	}

	if utils.FindPositionBySymbol(params.BaseMarketClass.APIData.GetPositions(), params.BaseMarketClass.GetSymbol().BrokerAPIName) != nil {
		log("There is an open position, no need to close any orders ...")
		params.BaseMarketClass.OnValidTradeSetup(onValidTradeSetupParams)
	} else {
		log("There isn't any open position. Closing orders first ...")
		params.BaseMarketClass.APIRetryFacade.CloseOrders(
			params.BaseMarketClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(params.BaseMarketClass.APIData.GetOrders(), params.BaseMarketClass.GetSymbol().BrokerAPIName)),
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
