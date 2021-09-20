package GER30

import (
	"TradingBot/src/constants"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies/strategies"
	"TradingBot/src/strategies/tickers/baseTickerClass"
	"TradingBot/src/strategies/tickers/interfaces"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"math"
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
	spreads           []float64
	averageSpread     float64

	mutex *sync.Mutex
}

func (s *Strategy) Parent() interfaces.BaseTickerClassInterface {
	return &s.BaseTickerClass
}

// Initialize ...
func (s *Strategy) Initialize() {
	s.BaseTickerClass.Initialize()

	s.mutex = &sync.Mutex{}
	s.BaseTickerClass.CandlesHandler.InitCandles(time.Now(), "")
	go s.BaseTickerClass.CheckNewestOpenedPositionSLandTP(
		&ResistanceBreakoutParams,
		&SupportBreakoutParams,
	)

	s.isReady = true
}

// DailyReset ...
func (s *Strategy) DailyReset() {
	s.BaseTickerClass.Initialize()

	s.isReady = false
	s.BaseTickerClass.CandlesHandler.InitCandles(time.Now(), "")
	s.isReady = true

	s.BaseTickerClass.SetPendingOrder(nil)
}

// OnReceiveMarketData ...
func (s *Strategy) OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.BaseTickerClass.OnReceiveMarketData(symbol, data)

	if !s.isReady {
		return
	}

	s.mutex.Lock()
	defer func() {
		s.mutex.Unlock()
	}()

	go s.updateAverageSpread()

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
	if data.Price != nil {
		// There is more or less a discrepancy of .8 between the price of ibroker and the price of fx:ger30 on tradingview
		var price = *data.Price + .8
		data.Price = &price
	}
	s.BaseTickerClass.CandlesHandler.UpdateCandles(data, s.BaseTickerClass.GetCurrentExecutionTime(), s.lastVolume)

	if s.lastCandlesAmount != len(s.BaseTickerClass.CandlesHandler.GetCandles()) {
		if !utils.IsNowWithinTradingHours(s.BaseTickerClass.GetSymbol()) {
			s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Doing nothing - Now it's not the time.")

			s.BaseTickerClass.APIRetryFacade.CloseOrders(
				s.BaseTickerClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseTickerClass.GetOrders(), s.BaseTickerClass.GetSymbol().BrokerAPIName)),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback: func() {
						s.BaseTickerClass.SetOrders(nil)
						s.BaseTickerClass.SetPendingOrder(nil)

						p := utils.FindPositionBySymbol(s.BaseTickerClass.GetPositions(), s.BaseTickerClass.GetSymbol().BrokerAPIName)
						if p != nil {
							s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Closing the open position ... "+utils.GetStringRepresentation(p))
							s.BaseTickerClass.APIRetryFacade.ClosePosition(
								p.Instrument,
								retryFacade.RetryParams{
									DelayBetweenRetries: 5 * time.Second,
									MaxRetries:          30,
									SuccessCallback:     func() { s.BaseTickerClass.SetPositions(nil) },
								},
							)
						}
					},
				})

			return
		}

		if s.averageSpread > s.BaseTickerClass.GetSymbol().MaxSpread {
			s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Closing working orders and doing nothing since the spread is very big -> "+utils.FloatToString(s.averageSpread, 0))
			s.BaseTickerClass.SetPendingOrder(nil)
			s.BaseTickerClass.APIRetryFacade.CloseOrders(
				s.BaseTickerClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseTickerClass.GetOrders(), s.BaseTickerClass.GetSymbol().BrokerAPIName)),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback:     func() { s.BaseTickerClass.SetOrders(nil) },
				},
			)
			return
		}

		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Calling supportBreakoutAnticipationStrategy")
		strategies.SupportBreakoutAnticipation(strategies.StrategyParams{
			BaseTickerClass:       &s.BaseTickerClass,
			TickerStrategyParams:  &SupportBreakoutParams,
			WithPendingOrders:     true,
			CloseOrdersOnBadTrend: true,
		})
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Calling resistanceBreakoutAnticipationStrategy")
		strategies.ResistanceBreakoutAnticipation(strategies.StrategyParams{
			BaseTickerClass:       &s.BaseTickerClass,
			TickerStrategyParams:  &ResistanceBreakoutParams,
			WithPendingOrders:     true,
			CloseOrdersOnBadTrend: false,
		})
	} else {
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Doing nothing - still same candle")
	}
}

func (s *Strategy) updateAverageSpread() {
	if s.lastAsk == nil || s.lastBid == nil {
		return
	}

	if len(s.spreads) == 1500 {
		s.spreads = s.spreads[1:]
	}

	s.spreads = append(s.spreads, math.Abs(*s.lastAsk-*s.lastBid))

	var sum float64
	for _, spread := range s.spreads {
		sum += spread
	}

	s.averageSpread = sum / float64(len(s.spreads))
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
			Name:           "GER30 Strategy",
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
