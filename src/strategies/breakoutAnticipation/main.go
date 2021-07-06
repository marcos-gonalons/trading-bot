package breakoutAnticipation

import (
	"TradingBot/src/constants"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/logger"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"math"
	"strconv"
	"sync"
	"time"

	funk "github.com/thoas/go-funk"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// MainStrategyName ...
const MainStrategyName = "Breakout Anticipation Strategy"

// Strategy ...
type Strategy struct {
	APIRetryFacade          retryFacade.Interface
	API                     api.Interface
	Logger                  logger.Interface
	CandlesHandler          candlesHandler.Interface
	HorizontalLevelsService horizontalLevels.Interface
	Mutex                   *sync.Mutex

	Name      string
	Symbol    types.Symbol
	Timeframe types.Timeframe

	currentExecutionTime time.Time
	lastCandlesAmount    int
	lastVolume           float64
	lastBid              *float64
	lastAsk              *float64
	spreads              []float64
	averageSpread        float64

	orders             []*api.Order
	currentBrokerQuote *api.Quote
	positions          []*api.Position
	state              *api.State

	pendingOrder    *api.Order
	currentPosition *api.Position

	modifyingPositionTimestamp int64

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

// GetSymbol ...
func (s *Strategy) GetSymbol() *types.Symbol {
	return &s.Symbol
}

// Initialize ...
func (s *Strategy) Initialize() {
	s.Mutex = &sync.Mutex{}

	s.CandlesHandler.InitCandles(time.Now())
	go s.checkOpenPositionSLandTP()

	s.isReady = true
}

// Reset ...
func (s *Strategy) Reset() {
	s.isReady = false
	s.CandlesHandler.InitCandles(time.Now())
	s.isReady = true
	s.pendingOrder = nil
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

		s.lastCandlesAmount = len(s.CandlesHandler.GetCandles())
		s.log(MainStrategyName, "Candles amount -> "+strconv.Itoa(s.lastCandlesAmount))
	}()

	s.log(MainStrategyName, "Updating candles... ")
	if data.Price != nil {
		// There is more or less a discrepancy of .8 between the price of ibroker and the price of fx:ger30 on tradingview
		var price = *data.Price + .8
		data.Price = &price
	}
	s.CandlesHandler.UpdateCandles(data, s.currentExecutionTime, s.lastVolume)

	if s.lastCandlesAmount != len(s.CandlesHandler.GetCandles()) {
		if !utils.IsNowWithinTradingHours(s.GetSymbol()) {
			s.log(MainStrategyName, "Doing nothing - Now it's not the time.")
			s.APIRetryFacade.CloseOrders(
				s.API.GetWorkingOrders(s.orders),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback: func() {
						s.orders = nil
						s.pendingOrder = nil

						p := s.getOpenPosition()
						if p != nil {
							s.log(MainStrategyName, "Closing the open position ... "+utils.GetStringRepresentation(p))
							s.APIRetryFacade.ClosePosition(
								p.Instrument,
								retryFacade.RetryParams{
									DelayBetweenRetries: 5 * time.Second,
									MaxRetries:          30,
									SuccessCallback:     func() { s.positions = nil },
								},
							)
						}
					},
				})

			return
		}

		if s.averageSpread > s.GetSymbol().MaxSpread {
			s.log(MainStrategyName, "Closing working orders and doing nothing since the spread is very big -> "+utils.FloatToString(s.averageSpread, 0))
			s.pendingOrder = nil
			s.APIRetryFacade.CloseOrders(
				s.API.GetWorkingOrders(s.orders),
				retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback:     func() { s.orders = nil },
				},
			)
			return
		}

		s.log(MainStrategyName, "Calling supportBreakoutAnticipationStrategy")
		s.supportBreakoutAnticipationStrategy(s.CandlesHandler.GetCandles())
		s.log(MainStrategyName, "Calling resistanceBreakoutAnticipationStrategy")
		s.resistanceBreakoutAnticipationStrategy(s.CandlesHandler.GetCandles())
	} else {
		s.log(MainStrategyName, "Doing nothing - still same candle")
	}
}

func (s *Strategy) checkOpenPositionSLandTP() {
	for {
		p := s.getOpenPosition()

		if p != nil && s.currentPosition == nil {
			s.currentPosition = p

			var tp string
			var sl string

			// todo: get the tp and sl accordingly
			if s.API.IsShortPosition(s.currentPosition) {
				tp = utils.FloatToString(float64(s.currentPosition.AvgPrice-34), s.GetSymbol().PriceDecimals)
				sl = utils.FloatToString(float64(s.currentPosition.AvgPrice+15), s.GetSymbol().PriceDecimals)
			} else {
				tp = utils.FloatToString(float64(s.currentPosition.AvgPrice+34), s.GetSymbol().PriceDecimals)
				sl = utils.FloatToString(float64(s.currentPosition.AvgPrice-24), s.GetSymbol().PriceDecimals)
			}
			s.APIRetryFacade.ModifyPosition(s.Symbol.BrokerAPIName, tp, sl, retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			})
		}

		if p == nil {
			s.currentPosition = nil
		}

		time.Sleep(5 * time.Second)
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
		s.log(MainStrategyName, "Save pending order called for side "+side)

		if s.getOpenPosition() != nil {
			s.log(MainStrategyName, "Can't save pending order since there is an open position")
			return
		}

		workingOrders := s.API.GetWorkingOrders(s.orders)

		if len(workingOrders) == 0 {
			s.log(MainStrategyName, "There aren't any working orders, doing nothing ...")
			return
		}

		var mainOrder *api.Order
		for _, workingOrder := range workingOrders {
			if workingOrder.Side == side && workingOrder.ParentID == nil {
				mainOrder = workingOrder
			}
		}

		if mainOrder == nil {
			s.log(MainStrategyName, "There isn't an active order for this side "+side)
			return
		}

		var validHalfHours []string
		if side == ibroker.LongSide {
			validHalfHours = s.GetSymbol().ValidTradingTimes.Longs.ValidHalfHours
		} else {
			validHalfHours = s.GetSymbol().ValidTradingTimes.Shorts.ValidHalfHours
		}

		if s.isExecutionTimeValid(
			[]string{},
			[]string{},
			validHalfHours,
		) {
			s.log(MainStrategyName, "No need to save the pending order since we are in the right time")
			return
		}

		// TODO: savingPendingOrderTimestamp

		s.log(MainStrategyName, "Closing the current order and saving it for the future, since now it's not the time for profitable trading.")
		s.log(MainStrategyName, "This is the current order -> "+utils.GetStringRepresentation(mainOrder))

		slOrder, tpOrder := s.getSlAndTpOrders(mainOrder.ID, workingOrders)

		if slOrder != nil {
			mainOrder.StopLoss = slOrder.StopPrice
		}
		if tpOrder != nil {
			mainOrder.TakeProfit = tpOrder.LimitPrice
		}

		if s.API.IsLimitOrder(mainOrder) {
			mainOrder.StopPrice = nil
		}
		if s.API.IsStopOrder(mainOrder) {
			mainOrder.LimitPrice = nil
		}

		s.APIRetryFacade.CloseOrders(
			s.API.GetWorkingOrders(s.orders),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					s.orders = nil
					s.pendingOrder = mainOrder
					s.log(MainStrategyName, "Closed all working orders correctly and pending order saved -> "+utils.GetStringRepresentation(s.pendingOrder))
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

		if s.API.IsLimitOrder(workingOrder) {
			tpOrder = workingOrder
		}
		if s.API.IsStopOrder(workingOrder) {
			slOrder = workingOrder
		}
	}

	return slOrder, tpOrder
}

func (s *Strategy) createPendingOrder(side string) {
	if s.pendingOrder.Side != side {
		return
	}

	p := s.getOpenPosition()
	if p != nil {
		s.log(MainStrategyName, "Can't create the pending order since there is an open position -> "+utils.GetStringRepresentation(p))
		return
	}

	go func(pendingOrder *api.Order) {
		var price float32
		if s.API.IsStopOrder(pendingOrder) {
			price = *pendingOrder.StopPrice
		} else {
			price = *pendingOrder.LimitPrice
		}

		candles := s.CandlesHandler.GetCandles()
		lastCompletedCandle := candles[len(candles)-2]
		s.log(MainStrategyName, "Last completed candle -> "+utils.GetStringRepresentation(lastCompletedCandle))

		if side == ibroker.LongSide {
			if price <= float32(lastCompletedCandle.Close) {
				s.log(MainStrategyName, "Price is lower than last completed candle.high - Can't create the pending order")
				return
			}
		} else {
			if price >= float32(lastCompletedCandle.Low) {
				s.log(MainStrategyName, "Price is greater than last completed candle.low - Can't create the pending order")
				return
			}
		}

		s.log(MainStrategyName, "Everything is good - Creating the pending order")
		s.APIRetryFacade.CreateOrder(
			pendingOrder,
			func() *api.Quote {
				return s.currentBrokerQuote
			},
			s.setStringValues,
			retryFacade.RetryParams{
				DelayBetweenRetries: 10 * time.Second,
				MaxRetries:          20,
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

	if s.modifyingPositionTimestamp == s.CandlesHandler.GetLastCandle().Timestamp {
		return
	}

	s.log(MainStrategyName, "Checking if the current position needs to have the SL adjusted... ")
	s.log(MainStrategyName, "Current position is "+utils.GetStringRepresentation(position))

	_, tpOrder := s.getSlAndTpOrdersForCurrentOpenPosition()

	if tpOrder == nil {
		s.log(MainStrategyName, "Take Profit order not found ...")
		return
	}

	shouldBeAdjusted := false
	if s.API.IsLongPosition(position) {
		shouldBeAdjusted = float64(*tpOrder.LimitPrice)-s.CandlesHandler.GetLastCandle().High < distanceToTp
	} else {
		shouldBeAdjusted = s.CandlesHandler.GetLastCandle().Low-float64(*tpOrder.LimitPrice) < distanceToTp
	}

	if shouldBeAdjusted {
		s.log(MainStrategyName, "The price is very close to the TP. Adjusting SL to break even ...")
		s.modifyingPositionTimestamp = s.CandlesHandler.GetLastCandle().Timestamp

		s.APIRetryFacade.ModifyPosition(
			s.GetSymbol().BrokerAPIName,
			utils.FloatToString(float64(*tpOrder.LimitPrice), 2),
			utils.FloatToString(float64(position.AvgPrice), 2),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			},
		)
	} else {
		s.log(MainStrategyName, "The price is not close to the TP yet. Doing nothing ...")
	}
}

// todo: should be API method probably
func (s *Strategy) getSlAndTpOrdersForCurrentOpenPosition() (
	slOrder *api.Order,
	tpOrder *api.Order,
) {
	for _, order := range s.orders {
		if order.Status != "working" {
			continue
		}
		if s.API.IsLimitOrder(order) {
			tpOrder = order
		}
		if s.API.IsStopOrder(order) {
			slOrder = order
		}
	}
	return
}

// todo: should be API method probably
func (s *Strategy) getWorkingOrderWithBracketOrders(side string, symbol string, orders []*api.Order) []*api.Order {
	var workingOrders []*api.Order

	for _, order := range s.orders {
		if order.Status != "working" || order.Side != side || order.Instrument != symbol || order.ParentID != nil {
			continue
		}

		workingOrders = append(workingOrders, order)
	}

	if len(workingOrders) > 0 {
		for _, order := range s.orders {
			if order.Status != "working" || order.ParentID == nil || *order.ParentID != workingOrders[0].ID {
				continue
			}

			workingOrders = append(workingOrders, order)
		}
	}

	return workingOrders
}

func (s *Strategy) setStringValues(order *api.Order) {
	currentAsk := utils.FloatToString(float64(s.currentBrokerQuote.Ask), s.GetSymbol().PriceDecimals)
	currentBid := utils.FloatToString(float64(s.currentBrokerQuote.Bid), s.GetSymbol().PriceDecimals)
	qty := utils.IntToString(int64(order.Qty))
	order.StringValues = &api.OrderStringValues{
		CurrentAsk: &currentAsk,
		CurrentBid: &currentBid,
		Qty:        &qty,
	}

	if s.API.IsLimitOrder(order) {
		limitPrice := utils.FloatToString(math.Round(float64(*order.LimitPrice)*10)/10, s.GetSymbol().PriceDecimals)
		order.StringValues.LimitPrice = &limitPrice
	} else {
		stopPrice := utils.FloatToString(math.Round(float64(*order.StopPrice)*10)/10, s.GetSymbol().PriceDecimals)
		order.StringValues.StopPrice = &stopPrice
	}
	if order.StopLoss != nil {
		stopLossPrice := utils.FloatToString(math.Round(float64(*order.StopLoss)*10)/10, s.GetSymbol().PriceDecimals)
		order.StringValues.StopLoss = &stopLossPrice
	}
	if order.TakeProfit != nil {
		takeProfitPrice := utils.FloatToString(math.Round(float64(*order.TakeProfit)*10)/10, s.GetSymbol().PriceDecimals)
		order.StringValues.TakeProfit = &takeProfitPrice
	}
}

func (s *Strategy) log(strategyName string, message string) {
	s.Logger.Log(strategyName + " - " + message)
}

func (s *Strategy) getOpenPosition() *api.Position {
	p := funk.Find(s.positions, func(p *api.Position) bool {
		return p.Instrument == s.GetSymbol().BrokerAPIName
	})

	if p == nil {
		return nil
	}

	return p.(*api.Position)
}

// GetStrategyInstance ...
func GetStrategyInstance(
	api api.Interface,
	apiRetryFacade retryFacade.Interface,
	logger logger.Interface,
) *Strategy {
	return &Strategy{
		API:            api,
		APIRetryFacade: apiRetryFacade,
		Logger:         logger,
		Name:           MainStrategyName,
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
	}
}
