package EURUSD

import (
	"TradingBot/src/markets/baseMarketClass"
	"TradingBot/src/markets/interfaces"
	"TradingBot/src/services/api"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
	loggerTypes "TradingBot/src/services/logger/types"
	"TradingBot/src/types"
	"strconv"
)

// Market ...
type Market struct {
	baseMarketClass.BaseMarketClass
}

// DailyReset ...
func (s *Market) DailyReset() {
	minCandles := 7 * 2 * 24
	totalCandles := len(s.CandlesHandler.GetCandles())

	s.Log("Total candles is " + strconv.Itoa(totalCandles) + " - min candles is " + strconv.Itoa(minCandles))
	if totalCandles < minCandles {
		s.Log("Not removing any candles yet")
		return
	}

	var candlesToRemove uint = 25
	s.Log("Removing old candles ... " + strconv.Itoa(int(candlesToRemove)))
	s.CandlesHandler.RemoveOldCandles(candlesToRemove)
}

func (s *Market) OnNewCandle() {
	s.OnNewCandle()

	// todo: call ema crossover strategy, longs and shorts
	/*
		s.Log("Calling resistanceBounce strategy")
		strategies.ResistanceBounce(strategies.StrategyParams{
			BaseMarketClass:       &s.BaseMarketClass,
			MarketStrategyParams:  &ResistanceBounceParams,
			WithPendingOrders:     false,
			CloseOrdersOnBadTrend: false,
		})

		s.Log("Calling supportBounce strategy")
		strategies.SupportBounce(strategies.StrategyParams{
			BaseMarketClass:       &s.BaseMarketClass,
			MarketStrategyParams:  &SupportBounceParams,
			WithPendingOrders:     false,
			CloseOrdersOnBadTrend: false,
		})
	*/
}

// GetMarketInstance ...
func GetMarketInstance(
	api api.Interface,
	apiData api.DataInterface,
	apiRetryFacade retryFacade.Interface,
	logger logger.Interface,
) interfaces.MarketInterface {
	market := &Market{}

	market.MarketData = types.MarketData{
		BrokerAPIName: ibroker.EURUSDSymbolName,
		SocketName:    "FX:EURUSD",
		PriceDecimals: 5,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           999999,
		LogType:             loggerTypes.EURUSD,
		MarketType:          "forex", // todo: move to constant. we have 'forex' and 'index' for now
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
		CandlesFileName:  "EURUSD-4H.csv",
		Rollover:         .7, // Only used in market replay command
		LongSetupParams:  &EMACrossoverLongParams,
		ShortSetupParams: &EMACrossoverShortParams,
		EurExchangeRate:  1,
	}

	market.API = api
	market.APIData = apiData
	market.APIRetryFacade = apiRetryFacade
	market.Logger = logger

	return market
}
