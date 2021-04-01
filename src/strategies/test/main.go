package test

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/logger"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/types"
	"strconv"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// TestStrategyName ...
const TestStrategyName = "Test Strategy"

// Strategy ...
type Strategy struct {
	APIRetryFacade          retryFacade.Interface
	API                     api.Interface
	Logger                  logger.Interface
	CandlesHandler          candlesHandler.Interface
	HorizontalLevelsService horizontalLevels.Interface
	Mutex                   *sync.Mutex

	Name            string
	SymbolForSocket string
	SymbolForAPI    string
	Timeframe       types.Timeframe

	currentExecutionTime time.Time
	lastCandlesAmount    int
	lastVolume           float64

	orders             []*api.Order
	currentBrokerQuote *api.Quote
	positions          []*api.Position
	state              *api.State

	isReady bool
}

// SetCandlesHandler ...
func (s *Strategy) SetCandlesHandler(candlesHandler candlesHandler.Interface) {
	s.CandlesHandler = candlesHandler
}

// SetHorizontalLevelsService ...
func (s *Strategy) SetHorizontalLevelsService(horizontalLevelsService horizontalLevels.Interface) {
	s.HorizontalLevelsService = horizontalLevelsService
}

// GetTimeframe ...
func (s *Strategy) GetTimeframe() *types.Timeframe {
	return &s.Timeframe
}

// Initialize ...
func (s *Strategy) Initialize() {
	isReady := false
	for !isReady {
		// now := time.Now()
		// currentMinutes := now.Format("04")
		// currentHour := now.Format("15")
		/**
			timeframe unit -> m
			timeframe value -> 1

			important

			for minutes > 1, the initial candle must start anywhere at :00, :03, :06, etc
			Otherwise it will be fucked up

			So let's say the bot starts at 02 minutes. Example with candles of 3m
			It will create a candle from 02 to to 05, next one from 05 to 08, etc. It would end at :59, and then it repeats
			But we want the candles to go from :00 to :00, no from :02 to :02

			Same happens with candles of 5, 10 or 15 minutes.


			for hours: does it really matter where it starts?
		**/
		time.Sleep(time.Second)
	}

	s.Mutex = &sync.Mutex{}

	s.CandlesHandler.InitCandles(time.Now())
	s.isReady = true
}

// GetSymbolForSocket ...
func (s *Strategy) GetSymbolForSocket() string {
	return s.SymbolForSocket
}

// GetSymbolForAPI ...
func (s *Strategy) GetSymbolForAPI() string {
	return s.SymbolForAPI
}

// Reset ...
func (s *Strategy) Reset() {
	s.isReady = false
	s.CandlesHandler.InitCandles(time.Now())
	s.isReady = true
}

// SetOrders ...
func (s *Strategy) SetOrders(orders []*api.Order) {
	s.orders = orders
}

// SetCurrentBrokerQuote ...
func (s *Strategy) SetCurrentBrokerQuote(quote *api.Quote) {
	s.currentBrokerQuote = quote
}

// SetPositions ...
func (s *Strategy) SetPositions(positions []*api.Position) {
	s.positions = positions
}

// SetState ...
func (s *Strategy) SetState(state *api.State) {
	s.state = state
}

// OnReceiveMarketData ...
func (s *Strategy) OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	if !s.isReady {
		return
	}

	s.Mutex.Lock()
	defer func() {
		s.Mutex.Unlock()
	}()

	s.currentExecutionTime = time.Now()

	defer func() {
		if data.Volume != nil {
			s.lastVolume = *data.Volume
		}

		s.lastCandlesAmount = len(s.CandlesHandler.GetCandles())
		s.log(TestStrategyName, "Candles amount -> "+strconv.Itoa(s.lastCandlesAmount))
	}()

	s.log(TestStrategyName, "Updating candles... ")
	s.CandlesHandler.UpdateCandles(data, s.currentExecutionTime, s.lastVolume)

	if s.lastCandlesAmount != len(s.CandlesHandler.GetCandles()) {
		s.log(TestStrategyName, "New candle - time to execute strategy")
	} else {
		s.log(TestStrategyName, "Doing nothing - still same candle")
	}
}

func (s *Strategy) log(strategyName string, message string) {
	s.Logger.Log(strategyName + " - " + message)
}

// GetStrategyInstance ...
func GetStrategyInstance(
	api api.Interface,
	apiRetryFacade retryFacade.Interface,
	logger logger.Interface,
) *Strategy {
	var symbolForSocket = "BINANCE:BTCUSD"
	var symbolForAPI = "__test__"
	return &Strategy{
		API:             api,
		APIRetryFacade:  apiRetryFacade,
		Logger:          logger,
		Name:            TestStrategyName,
		SymbolForSocket: symbolForSocket,
		SymbolForAPI:    symbolForAPI,
		Timeframe: types.Timeframe{
			Value: 3,
			Unit:  "m",
		},
	}
}
