package GER30

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
	"math"
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
}

func (s *Strategy) Parent() interfaces.BaseClassInterface {
	return &s.BaseClass
}

// Initialize ...
func (s *Strategy) Initialize() {
	s.BaseClass.Initialize()

	s.BaseClass.Mutex = &sync.Mutex{}

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

	s.BaseClass.Mutex.Lock()
	defer func() {
		s.BaseClass.Mutex.Unlock()
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
		s.BaseClass.Log(s.BaseClass.Name, "Save pending order called for side "+side)

		if utils.FindPositionBySymbol(s.BaseClass.GetPositions(), s.BaseClass.GetSymbol().BrokerAPIName) != nil {
			s.BaseClass.Log(s.BaseClass.Name, "Can't save pending order since there is an open position")
			return
		}

		workingOrders := s.BaseClass.API.GetWorkingOrders(s.BaseClass.GetOrders())

		if len(workingOrders) == 0 {
			s.BaseClass.Log(s.BaseClass.Name, "There aren't any working orders, doing nothing ...")
			return
		}

		var mainOrder *api.Order
		for _, workingOrder := range workingOrders {
			if workingOrder.Side == side && workingOrder.ParentID == nil {
				mainOrder = workingOrder
			}
		}

		if mainOrder == nil {
			s.BaseClass.Log(s.BaseClass.Name, "There isn't an active order for this side "+side)
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
			s.BaseClass.Log(s.BaseClass.Name, "No need to save the pending order since we are in the right time")
			return
		}

		s.BaseClass.Log(s.BaseClass.Name, "Closing the current order and saving it for the future, since now it's not the time for profitable trading.")
		s.BaseClass.Log(s.BaseClass.Name, "This is the current order -> "+utils.GetStringRepresentation(mainOrder))

		slOrder, tpOrder := s.getSlAndTpOrders(mainOrder.ID, workingOrders)

		if slOrder != nil {
			mainOrder.StopLoss = slOrder.StopPrice
		}
		if tpOrder != nil {
			mainOrder.TakeProfit = tpOrder.LimitPrice
		}

		if s.BaseClass.API.IsLimitOrder(mainOrder) {
			mainOrder.StopPrice = nil
		}
		if s.BaseClass.API.IsStopOrder(mainOrder) {
			mainOrder.LimitPrice = nil
		}

		s.BaseClass.APIRetryFacade.CloseOrders(
			s.BaseClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseClass.GetOrders(), s.BaseClass.GetSymbol().BrokerAPIName)),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					s.BaseClass.SetOrders(nil)
					s.pendingOrder = mainOrder
					s.BaseClass.Log(s.BaseClass.Name, "Closed all working orders correctly and pending order saved -> "+utils.GetStringRepresentation(s.pendingOrder))
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

		if s.BaseClass.API.IsLimitOrder(workingOrder) {
			tpOrder = workingOrder
		}
		if s.BaseClass.API.IsStopOrder(workingOrder) {
			slOrder = workingOrder
		}
	}

	return slOrder, tpOrder
}

func (s *Strategy) createPendingOrder(side string) {
	if s.pendingOrder.Side != side {
		return
	}

	p := utils.FindPositionBySymbol(s.BaseClass.GetPositions(), s.BaseClass.GetSymbol().BrokerAPIName)
	if p != nil {
		s.BaseClass.Log(s.BaseClass.Name, "Can't create the pending order since there is an open position -> "+utils.GetStringRepresentation(p))
		return
	}

	go func(pendingOrder *api.Order) {
		var price float32
		if s.BaseClass.API.IsStopOrder(pendingOrder) {
			price = *pendingOrder.StopPrice
		} else {
			price = *pendingOrder.LimitPrice
		}

		candles := s.BaseClass.CandlesHandler.GetCandles()
		lastCompletedCandle := candles[len(candles)-2]
		s.BaseClass.Log(s.BaseClass.Name, "Last completed candle -> "+utils.GetStringRepresentation(lastCompletedCandle))

		if side == ibroker.LongSide {
			if price <= float32(lastCompletedCandle.Close) {
				s.BaseClass.Log(s.BaseClass.Name, "Price is lower than last completed candle.high - Can't create the pending order")
				return
			}
		} else {
			if price >= float32(lastCompletedCandle.Low) {
				s.BaseClass.Log(s.BaseClass.Name, "Price is greater than last completed candle.low - Can't create the pending order")
				return
			}
		}

		s.BaseClass.Log(s.BaseClass.Name, "Everything is good - Creating the pending order")
		order := s.pendingOrder
		s.BaseClass.APIRetryFacade.CreateOrder(
			pendingOrder,
			func() *api.Quote {
				return s.BaseClass.GetCurrentBrokerQuote()
			},
			s.setStringValues,
			retryFacade.RetryParams{
				DelayBetweenRetries: 10 * time.Second,
				MaxRetries:          20,
				SuccessCallback: func(order *api.Order) func() {
					return func() {
						s.currentOrder = order
						s.BaseClass.Log(s.BaseClass.Name, "Pending order successfully created ... "+utils.GetStringRepresentation(s.currentOrder))
					}
				}(order),
			},
		)
	}(s.pendingOrder)

	s.pendingOrder = nil
}

func (s *Strategy) checkIfSLShouldBeMovedToBreakEven(
	distanceToTp float64,
	position *api.Position,
) {
	if distanceToTp <= 0 {
		return
	}

	s.BaseClass.Log(s.BaseClass.Name, "Checking if the current position needs to have the SL adjusted... ")
	s.BaseClass.Log(s.BaseClass.Name, "Current position is "+utils.GetStringRepresentation(position))

	_, tpOrder := s.BaseClass.API.GetBracketOrdersForOpenedPosition(position)

	if tpOrder == nil {
		s.BaseClass.Log(s.BaseClass.Name, "Take Profit order not found ...")
		return
	}

	shouldBeAdjusted := false
	if s.BaseClass.API.IsLongPosition(position) {
		shouldBeAdjusted = float64(*tpOrder.LimitPrice)-s.BaseClass.CandlesHandler.GetLastCandle().High < distanceToTp
	} else {
		shouldBeAdjusted = s.BaseClass.CandlesHandler.GetLastCandle().Low-float64(*tpOrder.LimitPrice) < distanceToTp
	}

	if shouldBeAdjusted {
		s.BaseClass.Log(s.BaseClass.Name, "The price is very close to the TP. Adjusting SL to break even ...")

		s.BaseClass.APIRetryFacade.ModifyPosition(
			s.BaseClass.GetSymbol().BrokerAPIName,
			utils.FloatToString(float64(*tpOrder.LimitPrice), 2),
			utils.FloatToString(float64(position.AvgPrice), 2),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			},
		)
	} else {
		s.BaseClass.Log(s.BaseClass.Name, "The price is not close to the TP yet. Doing nothing ...")
	}
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
	size := math.Floor((s.BaseClass.GetState().Equity*(params.RiskPercentage/100))/float64(params.StopLossDistance+1) + 1)
	if size == 0 {
		size = 1
	}

	order := &api.Order{
		CurrentAsk: &s.BaseClass.GetCurrentBrokerQuote().Ask,
		CurrentBid: &s.BaseClass.GetCurrentBrokerQuote().Bid,
		Instrument: s.BaseClass.GetSymbol().BrokerAPIName,
		StopPrice:  &float32Price,
		Qty:        float32(size),
		Side:       params.Side,
		StopLoss:   &stopLoss,
		TakeProfit: &takeProfit,
		Type:       ibroker.StopType,
	}

	s.BaseClass.Log(strategyName, params.Side+" order to be created -> "+utils.GetStringRepresentation(order))

	if utils.FindPositionBySymbol(s.BaseClass.GetPositions(), s.BaseClass.GetSymbol().BrokerAPIName) != nil {
		s.BaseClass.Log(strategyName, "There is an open position, saving the order for later ...")
		s.pendingOrder = order
		return
	}

	if !params.IsValidTime {
		s.BaseClass.Log(strategyName, "Now is not the time for opening any "+params.Side+" orders, saving it for later ...")
		s.pendingOrder = order
		return
	}

	s.BaseClass.APIRetryFacade.CreateOrder(
		order,
		func() *api.Quote {
			return s.BaseClass.GetCurrentBrokerQuote()
		},
		s.setStringValues,
		retryFacade.RetryParams{
			DelayBetweenRetries: 10 * time.Second,
			MaxRetries:          20,
			SuccessCallback: func(order *api.Order) func() {
				return func() {
					s.currentOrder = order
					s.BaseClass.Log(strategyName, "New order successfully created ... "+utils.GetStringRepresentation(s.currentOrder))
				}
			}(order),
		},
	)
}

func (s *Strategy) setStringValues(order *api.Order) {
	order.CurrentAsk = &s.BaseClass.GetCurrentBrokerQuote().Ask
	order.CurrentBid = &s.BaseClass.GetCurrentBrokerQuote().Bid

	utils.SetStringValues(order, s.BaseClass.GetSymbol(), s.BaseClass.API)
}

func (s *Strategy) checkOpenPositionSLandTP() {
	for {
		position := utils.FindPositionBySymbol(s.BaseClass.GetPositions(), s.BaseClass.GetSymbol().BrokerAPIName)

		if position != nil && s.currentPosition == nil {
			s.currentPosition = position

			var tp string
			var sl string
			var closePosition bool = false

			if s.BaseClass.API.IsShortPosition(s.currentPosition) {
				if s.BaseClass.API.IsStopOrder(s.currentOrder) && float64(*s.currentOrder.StopPrice-s.currentPosition.AvgPrice) > SupportBreakoutParams.MaxTradeExecutionPriceDifference {
					closePosition = true
				}
				tp = utils.FloatToString(float64(s.currentPosition.AvgPrice-SupportBreakoutParams.TakeProfitDistance), s.BaseClass.GetSymbol().PriceDecimals)
				sl = utils.FloatToString(float64(s.currentPosition.AvgPrice+SupportBreakoutParams.StopLossDistance), s.BaseClass.GetSymbol().PriceDecimals)
			} else {
				if s.BaseClass.API.IsStopOrder(s.currentOrder) && float64(s.currentPosition.AvgPrice-*s.currentOrder.StopPrice) > SupportBreakoutParams.MaxTradeExecutionPriceDifference {
					closePosition = true
				}
				tp = utils.FloatToString(float64(s.currentPosition.AvgPrice+ResistanceBreakoutParams.TakeProfitDistance), s.BaseClass.GetSymbol().PriceDecimals)
				sl = utils.FloatToString(float64(s.currentPosition.AvgPrice-ResistanceBreakoutParams.StopLossDistance), s.BaseClass.GetSymbol().PriceDecimals)
			}

			if closePosition {
				s.BaseClass.Log(s.BaseClass.Name, "Will immediately close the position since it was executed very far away from the stop price")
				s.BaseClass.Log(s.BaseClass.Name, "Order is "+utils.GetStringRepresentation(s.currentOrder))
				s.BaseClass.Log(s.BaseClass.Name, "Position is "+utils.GetStringRepresentation(s.currentPosition))

				workingOrders := s.BaseClass.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.BaseClass.GetOrders(), s.BaseClass.GetSymbol().BrokerAPIName))
				s.BaseClass.Log(s.BaseClass.Name, "Closing working orders first ... "+utils.GetStringRepresentation(workingOrders))

				s.BaseClass.APIRetryFacade.CloseOrders(
					workingOrders,
					retryFacade.RetryParams{
						DelayBetweenRetries: 5 * time.Second,
						MaxRetries:          30,
						SuccessCallback: func() {
							s.BaseClass.SetOrders(nil)
							s.pendingOrder = nil

							s.BaseClass.Log(s.BaseClass.Name, "Closed all orders. Closing the position now ... ")
							s.BaseClass.APIRetryFacade.ClosePosition(s.currentPosition.Instrument, retryFacade.RetryParams{
								DelayBetweenRetries: 5 * time.Second,
								MaxRetries:          20,
							})
						},
					})
			} else {
				s.BaseClass.Log(s.BaseClass.Name, "Modifying the SL and TP of the recently open position ... ")
				s.BaseClass.APIRetryFacade.ModifyPosition(s.BaseClass.GetSymbol().BrokerAPIName, tp, sl, retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          20,
				})
			}
		}

		if position == nil {
			s.currentPosition = nil
		}

		time.Sleep(5 * time.Second)
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
