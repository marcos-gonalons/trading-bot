package EURUSD

/*
import (
	"TradingBot/src/constants"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies/baseClass"
	"TradingBot/src/strategies/interfaces"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"strconv"
	"sync"
	"time"

	funk "github.com/thoas/go-funk"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// Strategy ...
type Strategy struct {
	BaseClass baseClass.BaseClass

	pendingOrder         *api.Order
	currentExecutionTime time.Time
	isReady              bool
	lastCandlesAmount    int
	lastVolume           float64
	lastBid              *float64
	lastAsk              *float64
	spreads              []float64
	averageSpread        float64
	currentPosition      *api.Position
	currentOrder         *api.Order

	mutex *sync.Mutex
}

func (s *Strategy) Parent() interfaces.BaseClassInterface {
	return &s.BaseClass
}

// Initialize ...
func (s *Strategy) Initialize() {
	s.BaseClass.Initialize()

	s.mutex = &sync.Mutex{}
	s.BaseClass.CandlesHandler.InitCandles(time.Now(), "")
	go s.checkOpenPositionSLandTP()

	s.isReady = true
}

// DailyReset ...
func (s *Strategy) DailyReset() {
	s.BaseClass.Initialize()

	s.isReady = false
	s.BaseClass.CandlesHandler.InitCandles(time.Now(), "")
	s.isReady = true
	s.pendingOrder = nil
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

	go s.updateAverageSpread()

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
	if data.Price != nil {
		// There is more or less a discrepancy of .8 between the price of ibroker and the price of fx:ger30 on tradingview
		var price = *data.Price + .8
		data.Price = &price
	}
	s.BaseClass.CandlesHandler.UpdateCandles(data, s.currentExecutionTime, s.lastVolume)

	if s.lastCandlesAmount != len(s.BaseClass.CandlesHandler.GetCandles()) {
		if !utils.IsNowWithinTradingHours(s.BaseClass.GetSymbol()) {
			s.BaseClass.Log(s.BaseClass.Name, "Doing nothing - Now it's not the time.")

			s.BaseClass.APIRetryFacade.CloseOrders(
				s.BaseClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseClass.GetOrders(), s.BaseClass.GetSymbol().BrokerAPIName)),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback: func() {
						s.BaseClass.SetOrders(nil)
						s.pendingOrder = nil

						p := utils.FindPositionBySymbol(s.BaseClass.GetPositions(), s.BaseClass.GetSymbol().BrokerAPIName)
						if p != nil {
							s.BaseClass.Log(s.BaseClass.Name, "Closing the open position ... "+utils.GetStringRepresentation(p))
							s.BaseClass.APIRetryFacade.ClosePosition(
								p.Instrument,
								retryFacade.RetryParams{
									DelayBetweenRetries: 5 * time.Second,
									MaxRetries:          30,
									SuccessCallback:     func() { s.BaseClass.SetPositions(nil) },
								},
							)
						}
					},
				})

			return
		}

		if s.averageSpread > s.BaseClass.GetSymbol().MaxSpread {
			s.BaseClass.Log(s.BaseClass.Name, "Closing working orders and doing nothing since the spread is very big -> "+utils.FloatToString(s.averageSpread, 0))
			s.pendingOrder = nil
			s.BaseClass.APIRetryFacade.CloseOrders(
				s.BaseClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseClass.GetOrders(), s.BaseClass.GetSymbol().BrokerAPIName)),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback:     func() { s.BaseClass.SetOrders(nil) },
				},
			)
			return
		}

		s.BaseClass.Log(s.BaseClass.Name, "Calling supportBreakoutAnticipationStrategy")
		s.supportBreakoutAnticipationStrategy(s.BaseClass.CandlesHandler.GetCandles())
		s.BaseClass.Log(s.BaseClass.Name, "Calling resistanceBreakoutAnticipationStrategy")
		s.resistanceBreakoutAnticipationStrategy(s.BaseClass.CandlesHandler.GetCandles())
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
					return s.BrokerAPIName == ibroker.GER30SymbolName
				},
			).(types.Symbol),
			Timeframe: types.Timeframe{
				Value: 1,
				Unit:  "m",
			},
		},
	}
}
/**/
