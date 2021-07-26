package EURUSD

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
	s.BaseClass.CandlesHandler.InitCandles(time.Now(), "EURUSD-1H.csv")
	go s.BaseClass.CheckNewestOpenedPositionSLandTP(
		&SupportBounceParams,
		&ResistanceBounceParams,
	)

	s.isReady = true
}

// DailyReset ...
func (s *Strategy) DailyReset() {
	// no need to do anything
}

// OnReceiveMarketData ...
func (s *Strategy) OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.BaseClass.OnReceiveMarketData(symbol, data)

	if !s.isReady {
		return
	}

	s.mutex.Lock()
	defer func() {
		s.mutex.Unlock()
	}()

	s.currentExecutionTime = time.Now()
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
			Name:           "EURUSD Strategy",
			Symbol: funk.Find(
				constants.Symbols,
				func(s types.Symbol) bool {
					return s.BrokerAPIName == ibroker.EURUSDSymbolName
				},
			).(types.Symbol),
			Timeframe: types.Timeframe{
				Value: 1,
				Unit:  "h",
			},
		},
	}
}
