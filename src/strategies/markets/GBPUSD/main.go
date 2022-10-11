package GBPUSD

import (
	"TradingBot/src/constants"
	"TradingBot/src/services/api"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies/markets/baseMarketClass"
	"TradingBot/src/strategies/markets/interfaces"
	"TradingBot/src/types"
	"strconv"
	"sync"
	"time"

	funk "github.com/thoas/go-funk"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// Strategy ...
type Strategy struct {
	BaseMarketClass baseMarketClass.BaseMarketClass

	isReady           bool
	lastCandlesAmount int
	lastVolume        float64
	lastBid           *float64
	lastAsk           *float64

	mutex *sync.Mutex
}

func (s *Strategy) Parent() interfaces.BaseMarketClassInterface {
	return &s.BaseMarketClass
}

// Initialize ...
func (s *Strategy) Initialize() {
	s.BaseMarketClass.Initialize()

	s.mutex = &sync.Mutex{}
	s.BaseMarketClass.CandlesHandler.InitCandles(time.Now(), "GBPUSD-4H.csv")
	go s.BaseMarketClass.CheckNewestOpenedPositionSLandTP(
		&EMACrossoverLongParams,
		&EMACrossoverShortParams,
	)

	// todo: get the usdeur quote
	s.BaseMarketClass.SetEurExchangeRate(.85)

	s.isReady = true
}

// DailyReset ...
func (s *Strategy) DailyReset() {
	// todo: get the usdeur quote
	s.BaseMarketClass.SetEurExchangeRate(.85)

	minCandles := 7 * 2 * 24
	totalCandles := len(s.BaseMarketClass.CandlesHandler.GetCandles())

	s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Total candles is "+strconv.Itoa(totalCandles)+" - min candles is "+strconv.Itoa(minCandles))
	if totalCandles < minCandles {
		s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Not removing any candles yet")
		return
	}

	var candlesToRemove uint = 25
	s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Removing old candles ... "+strconv.Itoa(int(candlesToRemove)))
	s.BaseMarketClass.CandlesHandler.RemoveOldCandles(candlesToRemove)
}

// OnReceiveMarketData ...
func (s *Strategy) OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.BaseMarketClass.OnReceiveMarketData(symbol, data)

	if !s.isReady {
		s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Not ready to process yet, doing nothing ...")
		return
	}

	s.BaseMarketClass.SetCurrentExecutionTime(time.Now())
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

		s.lastCandlesAmount = len(s.BaseMarketClass.CandlesHandler.GetCandles())
		s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Candles amount -> "+strconv.Itoa(s.lastCandlesAmount))
	}()

	s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Updating candles... ")
	s.BaseMarketClass.CandlesHandler.UpdateCandles(data, s.BaseMarketClass.GetCurrentExecutionTime(), s.lastVolume)

	if s.lastCandlesAmount != len(s.BaseMarketClass.CandlesHandler.GetCandles()) {
		s.OnNewCandle()
	} else {
		s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Doing nothing - still same candle")
	}
}

func (s *Strategy) OnNewCandle() {
	s.BaseMarketClass.OnNewCandle()

	// todo: call ema crossover strategy, longs and shorts
	/*
		s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Calling resistanceBounce strategy")
		strategies.ResistanceBounce(strategies.StrategyParams{
			BaseMarketClass:       &s.BaseMarketClass,
			MarketStrategyParams:  &ResistanceBounceParams,
			WithPendingOrders:     false,
			CloseOrdersOnBadTrend: false,
		})

		s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Calling supportBounce strategy")
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
) *Strategy {
	return &Strategy{
		BaseMarketClass: baseMarketClass.BaseMarketClass{
			API:            api,
			APIRetryFacade: apiRetryFacade,
			APIData:        apiData,
			Logger:         logger,
			Name:           "GBPUSD Strategy",
			Symbol: funk.Find(
				constants.Symbols,
				func(s types.Symbol) bool {
					return s.BrokerAPIName == ibroker.GBPUSDSymbolName
				},
			).(types.Symbol),
			Timeframe: types.Timeframe{
				Value: 4,
				Unit:  "h",
			},
		},
	}
}
