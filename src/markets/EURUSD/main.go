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
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// Market ...
type Market struct {
	baseMarketClass.BaseMarketClass

	isReady           bool
	lastCandlesAmount int
	lastVolume        float64
	lastBid           *float64
	lastAsk           *float64

	mutex *sync.Mutex
}

// Initialize ...
func (s *Market) Initialize() {
	s.BaseMarketClass.Initialize()

	s.mutex = &sync.Mutex{}
	s.CandlesHandler.InitCandles(time.Now(), "EURUSD-4H.csv")
	go s.CheckNewestOpenedPositionSLandTP(
		&EMACrossoverLongParams,
		&EMACrossoverShortParams,
	)

	s.SetEurExchangeRate(.85)

	s.isReady = true
}

// DailyReset ...
func (s *Market) DailyReset() {
	// todo: get the usdeur quote
	s.SetEurExchangeRate(.85)

	minCandles := 7 * 2 * 24
	totalCandles := len(s.CandlesHandler.GetCandles())

	s.Log(s.Name, "Total candles is "+strconv.Itoa(totalCandles)+" - min candles is "+strconv.Itoa(minCandles))
	if totalCandles < minCandles {
		s.Log(s.Name, "Not removing any candles yet")
		return
	}

	var candlesToRemove uint = 25
	s.Log(s.Name, "Removing old candles ... "+strconv.Itoa(int(candlesToRemove)))
	s.CandlesHandler.RemoveOldCandles(candlesToRemove)
}

// OnReceiveMarketData ...
func (s *Market) OnReceiveMarketData(data *tradingviewsocket.QuoteData) {
	s.OnReceiveMarketData(data)

	if !s.isReady {
		s.Log(s.Name, "Not ready to process yet, doing nothing ...")
		return
	}

	s.mutex.Lock()
	defer func() {
		s.mutex.Unlock()
	}()
	defer func() {
		if data.Volume != nil {
			s.lastVolume = *data.Volume
		}
		if data.Bid != nil {
			s.lastBid = data.Bid
		}
		if data.Ask != nil {
			s.lastAsk = data.Ask
		}

		s.lastCandlesAmount = len(s.CandlesHandler.GetCandles())
		s.Log(s.Name, "Candles amount -> "+strconv.Itoa(s.lastCandlesAmount))
	}()

	s.Log(s.Name, "Updating candles... ")
	s.CandlesHandler.UpdateCandles(data, s.lastVolume)

	if s.lastCandlesAmount != len(s.CandlesHandler.GetCandles()) {
		s.OnNewCandle()
	} else {
		s.Log(s.Name, "Doing nothing - still same candle")
	}
}

func (s *Market) OnNewCandle() {
	s.OnNewCandle()

	// todo: call ema crossover strategy, longs and shorts
	/*
		s.Log(s.Name, "Calling resistanceBounce strategy")
		strategies.ResistanceBounce(strategies.StrategyParams{
			BaseMarketClass:       &s.BaseMarketClass,
			MarketStrategyParams:  &ResistanceBounceParams,
			WithPendingOrders:     false,
			CloseOrdersOnBadTrend: false,
		})

		s.Log(s.Name, "Calling supportBounce strategy")
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
		Rollover:            .7,
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
	}

	market.API = api
	market.APIData = apiData
	market.APIRetryFacade = apiRetryFacade
	market.Logger = logger

	return market

	/*
		return &Strategy{
			API:            api,
			APIRetryFacade: apiRetryFacade,
			APIData:        apiData,
			Logger:         logger,
			Market: funk.Find(
				constants.Markets,
				func(s types.Market) bool {
					return s.BrokerAPIName == ibroker.EURUSDSymbolName
				},
			).(types.Market),
		}
	*/
}
