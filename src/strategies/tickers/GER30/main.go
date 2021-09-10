package GER30

import (
	"TradingBot/src/constants"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
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

	currentExecutionTime time.Time
	isReady              bool
	lastCandlesAmount    int
	lastVolume           float64
	lastBid              *float64
	lastAsk              *float64
	spreads              []float64
	averageSpread        float64

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

		s.lastCandlesAmount = len(s.BaseTickerClass.CandlesHandler.GetCandles())
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Candles amount -> "+strconv.Itoa(s.lastCandlesAmount))
	}()

	s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Updating candles... ")
	if data.Price != nil {
		// There is more or less a discrepancy of .8 between the price of ibroker and the price of fx:ger30 on tradingview
		var price = *data.Price + .8
		data.Price = &price
	}
	s.BaseTickerClass.CandlesHandler.UpdateCandles(data, s.currentExecutionTime, s.lastVolume)

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
		s.supportBreakoutAnticipationStrategy(s.BaseTickerClass.CandlesHandler.GetCandles())
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Calling resistanceBreakoutAnticipationStrategy")
		s.resistanceBreakoutAnticipationStrategy(s.BaseTickerClass.CandlesHandler.GetCandles())
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

func (s *Strategy) isExecutionTimeValid(
	validMonths []string,
	validWeekDays []string,
	validHalfHours []string,
) bool {
	if len(validMonths) > 0 {
		if !utils.IsInArray(s.currentExecutionTime.Format("January"), validMonths) {
			return false
		}
	}

	if len(validWeekDays) > 0 {
		if !utils.IsInArray(s.currentExecutionTime.Format("Monday"), validWeekDays) {
			return false
		}
	}

	if len(validHalfHours) > 0 {
		currentHour, currentMinutes := utils.GetCurrentTimeHourAndMinutes()
		if currentMinutes >= 30 {
			currentMinutes = 30
		} else {
			currentMinutes = 0
		}

		currentHourString := strconv.Itoa(currentHour)
		currentMinutesString := strconv.Itoa(currentMinutes)
		if len(currentMinutesString) == 1 {
			currentMinutesString += "0"
		}

		return utils.IsInArray(currentHourString+":"+currentMinutesString, validHalfHours)
	}

	return true
}

func (s *Strategy) savePendingOrder(side string) {
	go func() {
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Save pending order called for side "+side)

		if utils.FindPositionBySymbol(s.BaseTickerClass.GetPositions(), s.BaseTickerClass.GetSymbol().BrokerAPIName) != nil {
			s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Can't save pending order since there is an open position")
			return
		}

		workingOrders := s.BaseTickerClass.API.GetWorkingOrders(s.BaseTickerClass.GetOrders())

		if len(workingOrders) == 0 {
			s.BaseTickerClass.Log(s.BaseTickerClass.Name, "There aren't any working orders, doing nothing ...")
			return
		}

		var mainOrder *api.Order
		for _, workingOrder := range workingOrders {
			if workingOrder.Side == side && workingOrder.ParentID == nil {
				mainOrder = workingOrder
			}
		}

		if mainOrder == nil {
			s.BaseTickerClass.Log(s.BaseTickerClass.Name, "There isn't an active order for this side "+side)
			return
		}

		var validHalfHours []string
		if side == ibroker.LongSide {
			validHalfHours = ResistanceBreakoutParams.ValidTradingTimes.ValidHalfHours
		} else {
			validHalfHours = SupportBreakoutParams.ValidTradingTimes.ValidHalfHours
		}

		if s.isExecutionTimeValid(
			[]string{},
			[]string{},
			validHalfHours,
		) {
			s.BaseTickerClass.Log(s.BaseTickerClass.Name, "No need to save the pending order since we are in the right time")
			return
		}

		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Closing the current order and saving it for the future, since now it's not the time for profitable trading.")
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "This is the current order -> "+utils.GetStringRepresentation(mainOrder))

		slOrder, tpOrder := s.getSlAndTpOrders(mainOrder.ID, workingOrders)

		if slOrder != nil {
			mainOrder.StopLoss = slOrder.StopPrice
		}
		if tpOrder != nil {
			mainOrder.TakeProfit = tpOrder.LimitPrice
		}

		if s.BaseTickerClass.API.IsLimitOrder(mainOrder) {
			mainOrder.StopPrice = nil
		}
		if s.BaseTickerClass.API.IsStopOrder(mainOrder) {
			mainOrder.LimitPrice = nil
		}

		s.BaseTickerClass.APIRetryFacade.CloseOrders(
			s.BaseTickerClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseTickerClass.GetOrders(), s.BaseTickerClass.GetSymbol().BrokerAPIName)),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					s.BaseTickerClass.SetOrders(nil)
					s.BaseTickerClass.SetPendingOrder(mainOrder)
					s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Closed all working orders correctly and pending order saved -> "+utils.GetStringRepresentation(s.BaseTickerClass.GetPendingOrder()))
				},
			},
		)
	}()
}

func (s *Strategy) getSlAndTpOrders(
	parentID string,
	orders []*api.Order,
) (*api.Order, *api.Order) {
	var slOrder *api.Order
	var tpOrder *api.Order
	for _, workingOrder := range orders {
		if workingOrder.ParentID == nil || *workingOrder.ParentID != parentID {
			continue
		}

		if s.BaseTickerClass.API.IsLimitOrder(workingOrder) {
			tpOrder = workingOrder
		}
		if s.BaseTickerClass.API.IsStopOrder(workingOrder) {
			slOrder = workingOrder
		}
	}

	return slOrder, tpOrder
}

func (s *Strategy) createPendingOrder(side string) {
	if s.BaseTickerClass.GetPendingOrder().Side != side {
		return
	}

	p := utils.FindPositionBySymbol(s.BaseTickerClass.GetPositions(), s.BaseTickerClass.GetSymbol().BrokerAPIName)
	if p != nil {
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Can't create the pending order since there is an open position -> "+utils.GetStringRepresentation(p))
		return
	}

	go func(pendingOrder *api.Order) {
		var price float32
		if s.BaseTickerClass.API.IsStopOrder(pendingOrder) {
			price = *pendingOrder.StopPrice
		} else {
			price = *pendingOrder.LimitPrice
		}

		candles := s.BaseTickerClass.CandlesHandler.GetCandles()
		lastCompletedCandle := candles[len(candles)-2]
		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Last completed candle -> "+utils.GetStringRepresentation(lastCompletedCandle))

		if side == ibroker.LongSide {
			if price <= float32(lastCompletedCandle.Close) {
				s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Price is lower than last completed candle.high - Can't create the pending order")
				return
			}
		} else {
			if price >= float32(lastCompletedCandle.Low) {
				s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Price is greater than last completed candle.low - Can't create the pending order")
				return
			}
		}

		s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Everything is good - Creating the pending order")
		order := s.BaseTickerClass.GetPendingOrder()
		s.BaseTickerClass.APIRetryFacade.CreateOrder(
			pendingOrder,
			func() *api.Quote {
				return s.BaseTickerClass.GetCurrentBrokerQuote()
			},
			s.BaseTickerClass.SetStringValues,
			retryFacade.RetryParams{
				DelayBetweenRetries: 10 * time.Second,
				MaxRetries:          20,
				SuccessCallback: func(order *api.Order) func() {
					return func() {
						s.BaseTickerClass.SetCurrentOrder(order)
						s.BaseTickerClass.Log(s.BaseTickerClass.Name, "Pending order successfully created ... "+utils.GetStringRepresentation(s.BaseTickerClass.GetCurrentOrder()))
					}
				}(order),
			},
		)
	}(s.BaseTickerClass.GetPendingOrder())

	s.BaseTickerClass.SetPendingOrder(nil)
}

type OnValidTradeSetupParams struct {
	Price              float64
	StopLossDistance   float32
	TakeProfitDistance float32
	RiskPercentage     float64
	IsValidTime        bool
	Side               string
}

func (s *Strategy) onValidTradeSetup(params OnValidTradeSetupParams) {
	float32Price := float32(params.Price)

	var strategyName string
	var stopLoss float32
	var takeProfit float32

	if params.Side == ibroker.LongSide {
		strategyName = s.getResistanceBreakoutStrategyName()
		stopLoss = float32Price - float32(params.StopLossDistance)
		takeProfit = float32Price + float32(params.TakeProfitDistance)
	} else {
		strategyName = s.getSupportBreakoutStrategyName()
		stopLoss = float32Price + float32(params.StopLossDistance)
		takeProfit = float32Price - float32(params.TakeProfitDistance)
	}

	// TOOD: move the getsize to another function somewhere
	size := math.Floor((s.BaseTickerClass.GetState().Equity*(params.RiskPercentage/100))/float64(params.StopLossDistance+1) + 1)
	if size == 0 {
		size = 1
	}

	order := &api.Order{
		CurrentAsk: &s.BaseTickerClass.GetCurrentBrokerQuote().Ask,
		CurrentBid: &s.BaseTickerClass.GetCurrentBrokerQuote().Bid,
		Instrument: s.BaseTickerClass.GetSymbol().BrokerAPIName,
		StopPrice:  &float32Price,
		Qty:        float32(size),
		Side:       params.Side,
		StopLoss:   &stopLoss,
		TakeProfit: &takeProfit,
		Type:       ibroker.StopType,
	}

	s.BaseTickerClass.Log(strategyName, params.Side+" order to be created -> "+utils.GetStringRepresentation(order))

	if utils.FindPositionBySymbol(s.BaseTickerClass.GetPositions(), s.BaseTickerClass.GetSymbol().BrokerAPIName) != nil {
		s.BaseTickerClass.Log(strategyName, "There is an open position, saving the order for later ...")
		s.BaseTickerClass.SetPendingOrder(order)
		return
	}

	if !params.IsValidTime {
		s.BaseTickerClass.Log(strategyName, "Now is not the time for opening any "+params.Side+" orders, saving it for later ...")
		s.BaseTickerClass.SetPendingOrder(order)
		return
	}

	s.BaseTickerClass.APIRetryFacade.CreateOrder(
		order,
		func() *api.Quote {
			return s.BaseTickerClass.GetCurrentBrokerQuote()
		},
		s.BaseTickerClass.SetStringValues,
		retryFacade.RetryParams{
			DelayBetweenRetries: 10 * time.Second,
			MaxRetries:          20,
			SuccessCallback: func(order *api.Order) func() {
				return func() {
					s.BaseTickerClass.SetCurrentOrder(order)
					s.BaseTickerClass.Log(strategyName, "New order successfully created ... "+utils.GetStringRepresentation(s.BaseTickerClass.GetCurrentOrder()))
				}
			}(order),
		},
	)
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
