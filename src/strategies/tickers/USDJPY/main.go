package USDJPY

import (
	"TradingBot/src/constants"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies/baseClass"
	"TradingBot/src/strategies/interfaces"
	"TradingBot/src/types"
	"strconv"
	"sync"
	"time"

	funk "github.com/thoas/go-funk"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// Strategy ...
type Strategy struct {
	BaseClass baseClass.BaseClass

	currentExecutionTime time.Time
	isReady              bool
	lastCandlesAmount    int
	lastVolume           float64
	lastBid              *float64
	lastAsk              *float64

	mutex *sync.Mutex
}

func (s *Strategy) Parent() interfaces.BaseClassInterface {
	return &s.BaseClass
}

// Initialize ...
func (s *Strategy) Initialize() {
	s.BaseClass.Initialize()

	s.mutex = &sync.Mutex{}
	s.BaseClass.CandlesHandler.InitCandles(time.Now(), "USDJPY-1H.csv")
	go s.BaseClass.CheckNewestOpenedPositionSLandTP(
		&SupportBounceParams,
		&ResistanceBounceParams,
	)

	s.isReady = true
}

// DailyReset ...
func (s *Strategy) DailyReset() {
	minCandles := 7 * 2 * 24
	totalCandles := len(s.BaseClass.CandlesHandler.GetCandles())

	s.BaseClass.Log(s.BaseClass.Name, "Total candles is "+strconv.Itoa(totalCandles)+" - min candles is "+strconv.Itoa(minCandles))
	if totalCandles < minCandles {
		s.BaseClass.Log(s.BaseClass.Name, "Not removing any candles yet")
		return
	}

	var candlesToRemove uint = 25
	s.BaseClass.Log(s.BaseClass.Name, "Removing old candles ... "+strconv.Itoa(int(candlesToRemove)))
	s.BaseClass.CandlesHandler.RemoveOldCandles(candlesToRemove)
}

// OnReceiveMarketData ...
func (s *Strategy) OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.BaseClass.OnReceiveMarketData(symbol, data)

	if !s.isReady {
		s.BaseClass.Log(s.BaseClass.Name, "Not ready to process yet, doing nothing ...")
		return
	}

	s.currentExecutionTime = time.Now()
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

		s.lastCandlesAmount = len(s.BaseClass.CandlesHandler.GetCandles())
		s.BaseClass.Log(s.BaseClass.Name, "Candles amount -> "+strconv.Itoa(s.lastCandlesAmount))
	}()

	s.BaseClass.Log(s.BaseClass.Name, "Updating candles... ")
	s.BaseClass.CandlesHandler.UpdateCandles(data, s.currentExecutionTime, s.lastVolume)

	if s.lastCandlesAmount != len(s.BaseClass.CandlesHandler.GetCandles()) {
		s.BaseClass.Log(s.BaseClass.Name, "New candle has been added. Executing strategy code ...")
		// TODO: Code
	} else {
		s.BaseClass.Log(s.BaseClass.Name, "Doing nothing - still same candle")
	}
}

// GetStrategyInstance ...
func GetStrategyInstance(
	api api.Interface,
	apiRetryFacade retryFacade.Interface,
	logger logger.Interface,
) *Strategy {
	return &Strategy{
		BaseClass: baseClass.BaseClass{
			API:            api,
			APIRetryFacade: apiRetryFacade,
			Logger:         logger,
			Name:           "USDJPY Strategy",
			Symbol: funk.Find(
				constants.Symbols,
				func(s types.Symbol) bool {
					return s.BrokerAPIName == ibroker.USDJPYSymbolName
				},
			).(types.Symbol),
			Timeframe: types.Timeframe{
				Value: 1,
				Unit:  "h",
			},
		},
	}
}
