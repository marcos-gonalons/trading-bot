package horizontalLevelBreakout

import (
	"TradingBot/src/markets"
	"TradingBot/src/services"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/strategies"
	"TradingBot/src/utils"
	"time"
)

func ResistanceBreakoutAnticipation(params strategies.Params) {
	params.Market.Log("resistanceBreakoutAnticipation started")
	defer func() {
		params.Market.Log("resistanceBreakoutAnticipation ended")
	}()
	container := services.GetServicesContainer()

	err := strategies.OnBegin(params)
	if err != nil {
		params.Market.Log(err.Error() + utils.GetStringRepresentation(params.MarketStrategyParams))
		return
	}

	candles := params.CandlesHandler.GetCompletedCandles()
	level, err := container.HorizontalLevelsService.GetResistance(horizontalLevels.GetLevelParams{
		StartAt: int64(len(candles) - 1),
		CandlesAmountToBeConsideredHorizontalLevel: *params.MarketStrategyParams.CandlesAmountForHorizontalLevel,
		Candles:        candles,
		CandlesToCheck: 300,
	})

	if err != nil {
		errorMessage := "Not a good long setup yet -> " + err.Error()
		params.Market.Log(errorMessage)
		return
	}

	price := level.Candle.High - params.MarketStrategyParams.LimitAndStopOrderPriceOffset
	if price <= params.CandlesHandler.GetLastCompletedCandle().Close {
		params.Market.Log("Price is lower than the current ask, so we can't create the long order now. Price is -> " + utils.FloatToString(price, params.MarketData.PriceDecimals))
		return
	}

	params.Market.Log("Ok, we might have a long setup at price " + utils.FloatToString(price, params.MarketData.PriceDecimals))
	if !container.TrendsService.IsBullishTrend(
		params.MarketStrategyParams.TrendCandles,
		params.MarketStrategyParams.TrendDiff,
		params.CandlesHandler.GetCompletedCandles(),
	) {
		params.Market.Log("At the end it wasn't a good long setup, doing nothing ...")

		if params.MarketStrategyParams.CloseOrdersOnBadTrend && utils.FindPositionByMarket(container.APIData.GetPositions(), params.MarketData.BrokerAPIName) == nil {
			params.Market.Log("There isn't an open position, closing long orders ...")
			container.APIRetryFacade.CloseOrders(
				container.API.GetWorkingOrderWithBracketOrders(ibroker.LongSide, params.MarketData.BrokerAPIName, container.APIData.GetOrders()),
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
		Side:                 ibroker.LongSide,
		WithPendingOrders:    params.MarketStrategyParams.WithPendingOrders,
		OrderType:            ibroker.StopType,
		MinPositionSize:      params.MarketData.MinPositionSize,
		PositionSizeStrategy: params.MarketStrategyParams.PositionSizeStrategy,
	}

	if utils.FindPositionByMarket(container.APIData.GetPositions(), params.MarketData.BrokerAPIName) != nil {
		params.Market.Log("There is an open position, no need to close any orders ...")
		params.Market.OnValidTradeSetup(onValidTradeSetupParams)
	} else {
		params.Market.Log("There isn't any open position. Closing orders first ...")
		container.APIRetryFacade.CloseOrders(
			container.API.GetWorkingOrders(utils.FilterOrdersByMarket(container.APIData.GetOrders(), params.MarketData.BrokerAPIName)),
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
