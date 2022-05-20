package GER30

import (
	"TradingBot/src/constants"
	"TradingBot/src/services/api"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies/markets/baseMarketClass"
	"TradingBot/src/strategies/markets/interfaces"
	"TradingBot/src/strategies/strategies"
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
	BaseMarketClass baseMarketClass.BaseMarketClass

	isReady           bool
	lastCandlesAmount int
	lastVolume        float64
	lastBid           *float64
	lastAsk           *float64
	spreads           []float64
	averageSpread     float64

	mutex *sync.Mutex
}

func (s *Strategy) Parent() interfaces.BaseMarketClassInterface {
	return &s.BaseMarketClass
}

// Initialize ...
func (s *Strategy) Initialize() {
	s.BaseMarketClass.Initialize()

	s.mutex = &sync.Mutex{}
	s.BaseMarketClass.CandlesHandler.InitCandles(time.Now(), "")
	go s.BaseMarketClass.CheckNewestOpenedPositionSLandTP(
		&ResistanceBreakoutParams,
		&SupportBreakoutParams,
	)

	s.isReady = true
}

// DailyReset ...
func (s *Strategy) DailyReset() {
	s.BaseMarketClass.Initialize()

	s.isReady = false
	s.BaseMarketClass.CandlesHandler.InitCandles(time.Now(), "")
	s.isReady = true

	s.BaseMarketClass.SetPendingOrder(nil)
}

// OnReceiveMarketData ...
func (s *Strategy) OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.BaseMarketClass.OnReceiveMarketData(symbol, data)

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

		s.lastCandlesAmount = len(s.BaseMarketClass.CandlesHandler.GetCandles())
		s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Candles amount -> "+strconv.Itoa(s.lastCandlesAmount))
	}()

	s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Updating candles... ")
	if data.Price != nil {
		// There is more or less a discrepancy of .8 between the price of ibroker and the price of fx:ger30 on tradingview
		var price = *data.Price + .8
		data.Price = &price
	}
	s.BaseMarketClass.CandlesHandler.UpdateCandles(data, s.BaseMarketClass.GetCurrentExecutionTime(), s.lastVolume)

	if s.lastCandlesAmount != len(s.BaseMarketClass.CandlesHandler.GetCandles()) {
		s.OnNewCandle()
	} else {
		s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Doing nothing - still same candle")
	}
}

func (s *Strategy) OnNewCandle() {
	s.BaseMarketClass.OnNewCandle()

	if !utils.IsNowWithinTradingHours(s.BaseMarketClass.GetSymbol()) {
		s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Doing nothing - Now it's not the time.")

		s.BaseMarketClass.APIRetryFacade.CloseOrders(
			s.BaseMarketClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseMarketClass.APIData.GetOrders(), s.BaseMarketClass.GetSymbol().BrokerAPIName)),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					s.BaseMarketClass.SetPendingOrder(nil)

					p := utils.FindPositionBySymbol(s.BaseMarketClass.APIData.GetPositions(), s.BaseMarketClass.GetSymbol().BrokerAPIName)
					if p != nil {
						s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Closing the open position ... "+utils.GetStringRepresentation(p))
						s.BaseMarketClass.APIRetryFacade.ClosePosition(
							p.Instrument,
							retryFacade.RetryParams{
								DelayBetweenRetries: 5 * time.Second,
								MaxRetries:          30,
								SuccessCallback:     func() { s.BaseMarketClass.APIData.SetPositions(nil) },
							},
						)
					}
				},
			})

		return
	}

	if s.averageSpread > s.BaseMarketClass.GetSymbol().MaxSpread {
		s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Closing working orders and doing nothing since the spread is very big -> "+utils.FloatToString(s.averageSpread, 0))
		s.BaseMarketClass.SetPendingOrder(nil)
		s.BaseMarketClass.APIRetryFacade.CloseOrders(
			s.BaseMarketClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseMarketClass.APIData.GetOrders(), s.BaseMarketClass.GetSymbol().BrokerAPIName)),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
			},
		)
		return
	}

	s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Calling supportBreakoutAnticipationStrategy")
	strategies.SupportBreakoutAnticipation(strategies.StrategyParams{
		BaseMarketClass:       &s.BaseMarketClass,
		MarketStrategyParams:  &SupportBreakoutParams,
		WithPendingOrders:     true,
		CloseOrdersOnBadTrend: true,
	})
	s.BaseMarketClass.Log(s.BaseMarketClass.Name, "Calling resistanceBreakoutAnticipationStrategy")
	strategies.ResistanceBreakoutAnticipation(strategies.StrategyParams{
		BaseMarketClass:       &s.BaseMarketClass,
		MarketStrategyParams:  &ResistanceBreakoutParams,
		WithPendingOrders:     true,
		CloseOrdersOnBadTrend: false,
	})
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
