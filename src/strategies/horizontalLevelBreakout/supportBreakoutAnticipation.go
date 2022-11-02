package horizontalLevelBreakout

import (
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/strategies"
	"TradingBot/src/utils"
	"time"
)

func SupportBreakoutAnticipation(params strategies.Params) {
	params.Market.Log("supportBreakoutAnticipation started")
	defer func() {
		params.Market.Log("supportBreakoutAnticipation ended")
	}()

	err := strategies.OnBegin(params)
	if err != nil {
		params.Market.Log(err.Error() + utils.GetStringRepresentation(params.MarketStrategyParams))
		return
	}

	lastCompletedCandleIndex := len(params.CandlesHandler.GetCandles()) - 2
	price, err := params.Container.HorizontalLevelsService.GetSupportPrice(
		*params.MarketStrategyParams.CandlesAmountForHorizontalLevel,
		lastCompletedCandleIndex,
		params.CandlesHandler.GetCandles(),
	)

	if err != nil {
		errorMessage := "Not a good short setup yet -> " + err.Error()
		params.Market.Log(errorMessage)
		return
	}

	price = price + params.MarketStrategyParams.LimitAndStopOrderPriceOffset
	if price >= params.Market.GetCurrentBrokerQuote().Bid {
		params.Market.Log("Price is lower than the current ask, so we can't create the short order now. Price is -> " + utils.FloatToString(price, params.MarketData.PriceDecimals))
		params.Market.Log("Quote is -> " + utils.GetStringRepresentation(params.Market.GetCurrentBrokerQuote()))
		return
	}

	params.Market.Log("Ok, we might have a short setup at price " + utils.FloatToString(price, params.MarketData.PriceDecimals))
	if !params.Container.TrendsService.IsBearishTrend(
		params.MarketStrategyParams.TrendCandles,
		params.MarketStrategyParams.TrendDiff,
		params.CandlesHandler.GetCandles(),
		lastCompletedCandleIndex,
	) {
		params.Market.Log("At the end it wasn't a good short setup")
		if params.MarketStrategyParams.CloseOrdersOnBadTrend && utils.FindPositionByMarket(params.Container.APIData.GetPositions(), params.MarketData.BrokerAPIName) == nil {
			params.Market.Log("There isn't an open position, closing short orders ...")
			params.Container.APIRetryFacade.CloseOrders(
				params.Container.API.GetWorkingOrderWithBracketOrders(ibroker.ShortSide, params.MarketData.BrokerAPIName, params.Container.APIData.GetOrders()),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
				},
			)
		}
		return
	}

	onValidTradeSetupParams := markets.OnValidTradeSetupParams{
		Price:              price,
		StopLossDistance:   params.MarketStrategyParams.StopLossDistance,
		TakeProfitDistance: params.MarketStrategyParams.TakeProfitDistance,
		RiskPercentage:     params.MarketStrategyParams.RiskPercentage,
		IsValidTime: utils.IsExecutionTimeValid(
			time.Now(),
			params.MarketStrategyParams.ValidTradingTimes.ValidMonths,
			params.MarketStrategyParams.ValidTradingTimes.ValidWeekdays,
			params.MarketStrategyParams.ValidTradingTimes.ValidHalfHours,
		),
		Side:              ibroker.ShortSide,
		WithPendingOrders: params.MarketStrategyParams.WithPendingOrders,
		OrderType:         ibroker.StopType,
		MinPositionSize:   params.MarketData.MinPositionSize,
	}

	if utils.FindPositionByMarket(params.Container.APIData.GetPositions(), params.MarketData.BrokerAPIName) != nil {
		params.Market.Log("There is an open position, no need to close any orders ...")
		params.Market.OnValidTradeSetup(onValidTradeSetupParams)
	} else {
		params.Market.Log("There isn't any open position. Closing orders first ...")
		params.Container.APIRetryFacade.CloseOrders(
			params.Container.API.GetWorkingOrders(utils.FilterOrdersByMarket(params.Container.APIData.GetOrders(), params.MarketData.BrokerAPIName)),
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
