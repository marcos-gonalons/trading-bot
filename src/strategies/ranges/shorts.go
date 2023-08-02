package ranges

import (
	"TradingBot/src/markets"
	"TradingBot/src/services"
	"TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler/indicators/movingAverage"
	"TradingBot/src/services/technicalAnalysis/ranges"
	"TradingBot/src/strategies"
	"TradingBot/src/utils"
	"time"
)

func RangesShorts(params strategies.Params) {
	var log = func(m string) {
		params.Market.Log("RangesShorts | " + m)
	}

	log("RangesShorts started")
	defer func() {
		log("RangesShorts ended")
	}()

	err := strategies.OnBegin(params)
	if err != nil {
		log(err.Error() + utils.GetStringRepresentation(params.MarketStrategyParams))
		return
	}
	candles := params.CandlesHandler.GetCompletedCandles()
	lastCompletedCandleIndex := len(candles) - 1
	lastCompletedCandle := candles[lastCompletedCandleIndex]

	container := services.GetServicesContainer()

	openPosition := utils.FindPositionByMarket(container.APIData.GetPositions(), params.MarketData.BrokerAPIName)
	if openPosition != nil && container.API.IsShortPosition(openPosition) {
		log("There is an open position - doing nothing ...")
		return
	}

	if params.MarketStrategyParams.Ranges.TrendyOnly {
		if container.EmaService.IsPriceAboveEMA(lastCompletedCandle, movingAverage.BASE_EMA) {
			log("TrendyOnly flag is active and the current price is below the base EMA - Doing nothing ...")
			return
		}
	}

	r := container.RangesService.GetRange(ranges.GetRangeParams{
		Candles:             candles,
		LastCompletedCandle: lastCompletedCandle,
		StrategyParams:      *params.MarketStrategyParams,
	})

	if r == nil {
		log("A range wasn't found, doing nothing ...")
		return
	}
	log("A range was found! " + utils.GetStringRepresentation(r))

	resistancesAverage, supportsAverage := container.RangesService.GetAverages(r)
	log("Resistances average" + utils.GetStringRepresentation(resistancesAverage))
	log("Supports average" + utils.GetStringRepresentation(supportsAverage))

	orderPrice, err := getOrderPrice(
		*params.MarketStrategyParams,
		resistancesAverage,
		supportsAverage,
		lastCompletedCandle,
		constants.ShortSide,
	)
	log("Order price will be" + utils.GetStringRepresentation(orderPrice))
	if err != nil {
		log("An error occured while getting the order price " + err.Error())
		return
	}

	stopLoss := getStopLoss(
		*params.MarketStrategyParams,
		resistancesAverage,
		supportsAverage,
		orderPrice,
		constants.ShortSide,
	)
	log("Stop loss will be" + utils.GetStringRepresentation(stopLoss))
	if stopLoss <= orderPrice {
		log("Stop loss is lower than the order price, doing nothing ...")
		return
	}

	takeProfit := getTakeProfit(
		*params.MarketStrategyParams,
		resistancesAverage,
		supportsAverage,
		orderPrice,
		constants.ShortSide,
	)
	log("Take profit will be" + utils.GetStringRepresentation(takeProfit))
	if takeProfit >= orderPrice {
		log("Take profit is higher than the order price, doing nothing ...")
		return
	}

	var validMonths, validWeekdays, validHalfHours []string
	if params.MarketStrategyParams.ValidTradingTimes != nil {
		validMonths = params.MarketStrategyParams.ValidTradingTimes.ValidMonths
		validWeekdays = params.MarketStrategyParams.ValidTradingTimes.ValidWeekdays
		validHalfHours = params.MarketStrategyParams.ValidTradingTimes.ValidHalfHours
	}
	onValidTradeSetupParams := markets.OnValidTradeSetupParams{
		Price:              orderPrice,
		StopLossDistance:   stopLoss - orderPrice,
		TakeProfitDistance: orderPrice - takeProfit,
		RiskPercentage:     params.MarketStrategyParams.RiskPercentage,
		IsValidTime: utils.IsExecutionTimeValid(
			time.Now(),
			validMonths,
			validWeekdays,
			validHalfHours,
		),
		Side:                 constants.ShortSide,
		WithPendingOrders:    params.MarketStrategyParams.WithPendingOrders,
		OrderType:            params.MarketStrategyParams.Ranges.OrderType,
		MinPositionSize:      params.MarketData.MinPositionSize,
		PositionSizeStrategy: params.MarketStrategyParams.PositionSizeStrategy,
	}

	log("We have a valid setup, closing existing orders first ...")
	container.APIRetryFacade.CloseOrders(
		container.API.GetWorkingOrders(utils.FilterOrdersByMarket(container.APIData.GetOrders(), params.MarketData.BrokerAPIName)),
		retryFacade.RetryParams{
			DelayBetweenRetries: 5 * time.Second,
			MaxRetries:          30,
			SuccessCallback: func() {
				log("Orders closed, calling OnValidTradeSetup ...")
				params.Market.OnValidTradeSetup(onValidTradeSetupParams)
			},
		},
	)
}
