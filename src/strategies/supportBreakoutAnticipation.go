package strategies

import (
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/utils"
	"time"
)

// SupportBreakoutAnticipation ...
func SupportBreakoutAnticipation(params StrategyParams) {
	params.Market.Log("supportBreakoutAnticipation started")
	defer func() {
		params.Market.Log("supportBreakoutAnticipation ended")
	}()

	validMonths := params.MarketStrategyParams.ValidTradingTimes.ValidMonths
	validWeekdays := params.MarketStrategyParams.ValidTradingTimes.ValidWeekdays
	validHalfHours := params.MarketStrategyParams.ValidTradingTimes.ValidHalfHours

	now := time.Now()
	if !utils.IsExecutionTimeValid(now, validMonths, []string{}, []string{}) || !utils.IsExecutionTimeValid(now, []string{}, validWeekdays, []string{}) {
		params.Market.Log("Today it's not the day for support breakout anticipation for " + params.MarketData.SocketName)
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
			params.Market.SavePendingOrder(ibroker.ShortSide, params.MarketStrategyParams.ValidTradingTimes)
		} else {
			if params.Market.GetPendingOrder() != nil {
				params.Market.CreatePendingOrder(ibroker.ShortSide)
			}
			params.Market.SetPendingOrder(nil)
		}
	}

	p := utils.FindPositionByMarket(params.APIData.GetPositions(), params.MarketData.BrokerAPIName)
	if p != nil && p.Side == ibroker.ShortSide {
		params.Market.CheckIfSLShouldBeAdjusted(params.MarketStrategyParams, p)
		params.Market.CheckOpenPositionTTL(params.MarketStrategyParams, p)
	}

	lastCompletedCandleIndex := len(params.CandlesHandler.GetCandles()) - 2
	price, err := params.HorizontalLevelsService.GetSupportPrice(*params.MarketStrategyParams.CandlesAmountForHorizontalLevel, lastCompletedCandleIndex)

	if err != nil {
		errorMessage := "Not a good short setup yet -> " + err.Error()
		params.Market.Log(errorMessage)
		return
	}

	price = price + params.MarketStrategyParams.LimitAndStopOrderPriceOffset
	if price >= float64(params.Market.GetCurrentBrokerQuote().Bid) {
		params.Market.Log("Price is lower than the current ask, so we can't create the short order now. Price is -> " + utils.FloatToString(price, params.MarketData.PriceDecimals))
		params.Market.Log("Quote is -> " + utils.GetStringRepresentation(params.Market.GetCurrentBrokerQuote()))
		return
	}

	params.Market.Log("Ok, we might have a short setup at price " + utils.FloatToString(price, params.MarketData.PriceDecimals))
	if !params.TrendsService.IsBearishTrend(
		params.MarketStrategyParams.TrendCandles,
		params.MarketStrategyParams.TrendDiff,
		params.CandlesHandler.GetCandles(),
		lastCompletedCandleIndex,
	) {
		params.Market.Log("At the end it wasn't a good short setup")
		if params.MarketStrategyParams.CloseOrdersOnBadTrend && utils.FindPositionByMarket(params.APIData.GetPositions(), params.MarketData.BrokerAPIName) == nil {
			params.Market.Log("There isn't an open position, closing short orders ...")
			params.APIRetryFacade.CloseOrders(
				params.API.GetWorkingOrderWithBracketOrders(ibroker.ShortSide, params.MarketData.BrokerAPIName, params.APIData.GetOrders()),
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
		IsValidTime:        isValidTimeToOpenAPosition,
		Side:               ibroker.ShortSide,
		WithPendingOrders:  params.MarketStrategyParams.WithPendingOrders,
		OrderType:          ibroker.StopType,
		MinPositionSize:    params.MarketStrategyParams.MinPositionSize,
	}

	if utils.FindPositionByMarket(params.APIData.GetPositions(), params.MarketData.BrokerAPIName) != nil {
		params.Market.Log("There is an open position, no need to close any orders ...")
		params.Market.OnValidTradeSetup(onValidTradeSetupParams)
	} else {
		params.Market.Log("There isn't any open position. Closing orders first ...")
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
