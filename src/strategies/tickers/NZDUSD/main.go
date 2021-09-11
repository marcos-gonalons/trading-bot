package NZDUSD

import (
	"TradingBot/src/constants"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies/tickers/baseTickerClass"
	"TradingBot/src/strategies/tickers/interfaces"
	"TradingBot/src/types"
	"strconv"
	"sync"
	"time"

	funk "github.com/thoas/go-funk"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// Strategy ...
type Strategy struct {
	BaseTickerClass baseTickerClass.BaseTickerClass

	isReady           bool
	lastCandlesAmount int
	lastVolume        float64
	lastBid           *float64
	lastAsk           *float64

	mutex *sync.Mutex
}

func (s *Strategy) Parent() interfaces.BaseTickerClassInterface {
	return &s.BaseTickerClass
}

// Initialize ...
func (s *Strategy) Initialize() {
	s.BaseTickerClass.Initialize()

	s.mutex = &sync.Mutex{}
	s.BaseTickerClass.CandlesHandler.InitCandles(time.Now(), "NZDUSD-1H.csv")
	go s.BaseTickerClass.CheckNewestOpenedPositionSLandTP(
		&SupportBounceParams,
		nil,
	)

	s.isReady = true
}

// DailyReset ...
func (s *Strategy) DailyReset() {
	minCandles := 7 * 2 * 24
	totalCandles := len(s.BaseTickerClass.CandlesHandler.GetCandles())

	s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Total candles is "+strconv.Itoa(totalCandles)+" - min candles is "+strconv.Itoa(minCandles))
	if totalCandles < minCandles {
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Not removing any candles yet")
		return
	}

	var candlesToRemove uint = 25
	s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Removing old candles ... "+strconv.Itoa(int(candlesToRemove)))
	s.BaseTickerClass.CandlesHandler.RemoveOldCandles(candlesToRemove)
}

// OnReceiveMarketData ...
func (s *Strategy) OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.BaseTickerClass.OnReceiveMarketData(symbol, data)

	if !s.isReady {
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Not ready to process yet, doing nothing ...")
		return
	}

	s.BaseTickerClass.SetCurrentExecutionTime(time.Now())
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

		s.lastCandlesAmount = len(s.BaseTickerClass.CandlesHandler.GetCandles())
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Candles amount -> "+strconv.Itoa(s.lastCandlesAmount))
	}()

	s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Updating candles... ")
	s.BaseTickerClass.CandlesHandler.UpdateCandles(data, s.BaseTickerClass.GetCurrentExecutionTime(), s.lastVolume)

	if s.lastCandlesAmount != len(s.BaseTickerClass.CandlesHandler.GetCandles()) {
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "New candle has been added. Executing strategy code ...")
		// TODO: Code
	} else {
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Doing nothing - still same candle")
	}
}

// GetStrategyInstance ...
func GetStrategyInstance(
	api api.Interface,
	apiRetryFacade retryFacade.Interface,
	logger logger.Interface,
) *Strategy {
	return &Strategy{
		BaseTickerClass: baseTickerClass.BaseTickerClass{
			API:            api,
			APIRetryFacade: apiRetryFacade,
			Logger:         logger,
			Name:           "NZDUSD Strategy",
			Symbol: funk.Find(
				constants.Symbols,
				func(s types.Symbol) bool {
					return s.BrokerAPIName == ibroker.NZDUSDSymbolName
				},
			).(types.Symbol),
			Timeframe: types.Timeframe{
				Value: 1,
				Unit:  "h",
			},
		},
	}
}
