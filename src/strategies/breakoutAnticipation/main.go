package breakoutAnticipation

import (
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

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper"
)

// MainStrategyName ...
const MainStrategyName = "Breakout Anticipation Strategy"

// Strategy ...
type Strategy struct {
	APIRetryFacade          retryFacade.Interface
	Logger                  logger.Interface
	CandlesHandler          candlesHandler.Interface
	HorizontalLevelsService horizontalLevels.Interface
	Mutex                   *sync.Mutex

	Name      string
	Symbol    string
	Timeframe types.Timeframe

	currentExecutionTime  time.Time
	previousExecutionTime time.Time
	lastVolume            float64
	lastBid               *float64
	lastAsk               *float64
	spreads               []float64
	averageSpread         float64

	orders             []*api.Order
	currentBrokerQuote *api.Quote
	positions          []*api.Position
	state              *api.State

	pendingOrder    *api.Order
	currentPosition *api.Position

	creatingOrderTimestamp     int64
	modifyingOrderTimestamp    int64
	modifyingPositionTimestamp int64
	closingOrdersTimestamp     int64
}

// SetCandlesHandler ...
func (s *Strategy) SetCandlesHandler(candlesHandler candlesHandler.Interface) {
	s.CandlesHandler = candlesHandler
}

// SetHorizontalLevelsService ...
func (s *Strategy) SetHorizontalLevelsService(horizontalLevelsService horizontalLevels.Interface) {
	s.HorizontalLevelsService = horizontalLevelsService
}

// Initialize ...
func (s *Strategy) Initialize() {
	s.Mutex = &sync.Mutex{}

	s.CandlesHandler.InitCandles()
	go s.checkOpenPositionSLandTP()
}

// Reset ...
func (s *Strategy) Reset() {
	s.CandlesHandler.InitCandles()
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
	s.Mutex.Lock()
	defer func() {
		s.Mutex.Unlock()
	}()

	go s.updateAverageSpread()

	s.currentExecutionTime = time.Now()

	defer func() {
		s.previousExecutionTime = s.currentExecutionTime

		if data.Volume != nil {
			s.lastVolume = *data.Volume
		}
		if data.Bid != nil {
			s.lastBid = data.Bid
		}
		if data.Ask != nil {
			s.lastAsk = data.Ask
		}
	}()

	s.log(MainStrategyName, "Updating candles... ")
	s.CandlesHandler.UpdateCandles(data, s.currentExecutionTime, s.previousExecutionTime, s.lastVolume)

	if len(s.CandlesHandler.GetCandles()) == 0 {
		return
	}

	if s.isCurrentTimeOutsideTradingHours() {
		s.log(MainStrategyName, "Doing nothing - Now it's not the time.")
		s.APIRetryFacade.CloseAllWorkingOrders(retryFacade.RetryParams{
			DelayBetweenRetries: 5 * time.Second,
			MaxRetries:          30,
			SuccessCallback: func() {
				s.orders = nil
				s.pendingOrder = nil
				s.APIRetryFacade.ClosePositions(retryFacade.RetryParams{
					DelayBetweenRetries: 5 * time.Second,
					MaxRetries:          30,
					SuccessCallback:     func() { s.positions = nil },
				})
			},
		})

		return
	}

	if s.averageSpread > 3 {
		/**
			Todo:
			Do not create the order if the spread is big, but still save the pending orders for the future
			Maybe when the time is right, the spread will be ok and the order can be created.

			So when it's time to create an order
			if spread  big, no no
			else do it as it does it right now
		**/
		s.log(MainStrategyName, "Doing nothing since the spread is very big -> "+utils.FloatToString(s.averageSpread, 0))
		s.pendingOrder = nil
		s.APIRetryFacade.CloseAllWorkingOrders(retryFacade.RetryParams{
			DelayBetweenRetries: 5 * time.Second,
			MaxRetries:          30,
			SuccessCallback:     func() { s.orders = nil },
		})
		return
	}

	if len(s.CandlesHandler.GetCandles()) < 2 {
		return
	}

	s.resistanceBreakoutAnticipationStrategy(s.CandlesHandler.GetCandles())
	s.supportBreakoutAnticipationStrategy(s.CandlesHandler.GetCandles())
}

func (s *Strategy) checkOpenPositionSLandTP() {
	for {
		if len(s.positions) > 0 && s.currentPosition == nil {
			s.currentPosition = s.positions[0]
			var tp string
			var sl string
			if s.currentPosition.Side == "sell" {
				tp = utils.FloatToString(float64(s.currentPosition.AvgPrice-27), 1)
				sl = utils.FloatToString(float64(s.currentPosition.AvgPrice+12), 1)
			} else {
				tp = utils.FloatToString(float64(s.currentPosition.AvgPrice+27), 1)
				sl = utils.FloatToString(float64(s.currentPosition.AvgPrice-12), 1)
			}
			s.APIRetryFacade.ModifyPosition(s.Symbol, tp, sl, retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			})
		}
		if len(s.positions) == 0 {
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

func (s *Strategy) isCurrentTimeOutsideTradingHours() bool {
	currentHour, currentMinutes := s.getCurrentTimeHourAndMinutes()
	return (currentHour < 7) || (currentHour > 21) || (currentHour == 21 && currentMinutes > 57)
}

func (s *Strategy) getCurrentTimeHourAndMinutes() (int, int) {
	currentHour, _ := strconv.Atoi(s.currentExecutionTime.Format("15"))
	currentMinutes, _ := strconv.Atoi(s.currentExecutionTime.Format("04"))

	return currentHour, currentMinutes
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
		currentHour, currentMinutes := s.getCurrentTimeHourAndMinutes()
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
	// TODO: GetWorkingOrders should be an API method
	workingOrders := utils.GetWorkingOrders(s.orders)

	if len(workingOrders) == 0 {
		return
	}

	var mainOrder *api.Order
	for _, workingOrder := range workingOrders {
		if workingOrder.Side == side && workingOrder.ParentID == nil {
			mainOrder = workingOrder
		}
	}

	if mainOrder == nil {
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

	if mainOrder.Type == "limit" {
		mainOrder.StopPrice = nil
	}
	if mainOrder.Type == "stop" {
		mainOrder.LimitPrice = nil
	}

	s.APIRetryFacade.CloseAllWorkingOrders(retryFacade.RetryParams{
		DelayBetweenRetries: 5 * time.Second,
		MaxRetries:          30,
		SuccessCallback: func() {
			s.orders = nil
			s.pendingOrder = mainOrder
			s.log(MainStrategyName, "Closed all working orders correctly and pending order saved -> "+utils.GetStringRepresentation(s.pendingOrder))
		},
	})
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

		if workingOrder.Type == "limit" {
			tpOrder = workingOrder
		}
		if workingOrder.Type == "stop" {
			slOrder = workingOrder
		}
	}

	return slOrder, tpOrder
}

func (s *Strategy) createPendingOrder() {
	s.log(MainStrategyName, "Trying to create the pending order ..."+utils.GetStringRepresentation(s.pendingOrder))

	// todo: creatependingordertimestamp

	s.APIRetryFacade.CreateOrder(
		s.pendingOrder,
		func() *api.Quote {
			return s.currentBrokerQuote
		},
		s.setStringValues,
		retryFacade.RetryParams{
			DelayBetweenRetries: 10 * time.Second,
			MaxRetries:          20,
		},
	)
	s.pendingOrder = nil
}

func (s *Strategy) checkIfSLShouldBeMovedToBreakEven(distanceToTp float64, side string) {
	if s.modifyingPositionTimestamp == s.CandlesHandler.GetLastCandle().Timestamp {
		return
	}

	// todo: fx:ger30 price is slightly different to ibroker:ger30 price
	// I already make up for this difference when selecting the price of an order.
	// For example, for support breakout anticipation, I do price := s.candles[i].Low + 3
	// meanwhile, for resistance breakoy anticipation, I do s.candles[i].High - 1
	// So I must take that into consideration here when checking if I should move the SL to break even.

	position := s.positions[0]
	if position.Side != side {
		return
	}

	_, tpOrder := s.getSlAndTpOrdersForCurrentOpenPosition()

	if tpOrder == nil {
		return
	}

	shouldBeAdjusted := false
	if side == "buy" {
		shouldBeAdjusted = float64(*tpOrder.LimitPrice)-s.CandlesHandler.GetLastCandle().High < distanceToTp
	} else {
		shouldBeAdjusted = s.CandlesHandler.GetLastCandle().Low-float64(*tpOrder.LimitPrice) < distanceToTp
	}

	if shouldBeAdjusted {
		s.log(MainStrategyName, "The trade is very close to the TP. Adjusting SL to break even ...")
		s.modifyingPositionTimestamp = s.CandlesHandler.GetLastCandle().Timestamp

		s.APIRetryFacade.ModifyPosition(
			ibroker.GER30SymbolName,
			utils.FloatToString(float64(*tpOrder.LimitPrice), 2),
			utils.FloatToString(float64(position.AvgPrice), 2),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			},
		)
	}
}

func (s *Strategy) getSlAndTpOrdersForCurrentOpenPosition() (
	slOrder *api.Order,
	tpOrder *api.Order,
) {
	for _, order := range s.orders {
		if order.Status != "working" {
			continue
		}
		if order.Type == "limit" {
			tpOrder = order
		}
		if order.Type == "stop" {
			slOrder = order
		}
	}
	return
}

func (s *Strategy) setStringValues(order *api.Order) {
	// TODO: The decimals amount depends on the symbol. For ger 30, it uses 1 decimal, for SPX500, it uses 2.
	currentAsk := utils.FloatToString(float64(s.currentBrokerQuote.Ask), 1)
	currentBid := utils.FloatToString(float64(s.currentBrokerQuote.Bid), 1)
	qty := utils.IntToString(int64(order.Qty))
	order.StringValues = &api.OrderStringValues{
		CurrentAsk: &currentAsk,
		CurrentBid: &currentBid,
		Qty:        &qty,
	}

	if order.Type == "limit" {
		limitPrice := utils.FloatToString(math.Round(float64(*order.LimitPrice)*10)/10, 1)
		order.StringValues.LimitPrice = &limitPrice
	} else {
		stopPrice := utils.FloatToString(math.Round(float64(*order.StopPrice)*10)/10, 1)
		order.StringValues.StopPrice = &stopPrice
	}
	if order.StopLoss != nil {
		stopLossPrice := utils.FloatToString(math.Round(float64(*order.StopLoss)*10)/10, 1)
		order.StringValues.StopLoss = &stopLossPrice
	}
	if order.TakeProfit != nil {
		takeProfitPrice := utils.FloatToString(math.Round(float64(*order.TakeProfit)*10)/10, 1)
		order.StringValues.TakeProfit = &takeProfitPrice
	}
}

func (s *Strategy) log(strategyName string, message string) {
	s.Logger.Log(strategyName + " - " + message)
}

// GetStrategyInstance ...
func GetStrategyInstance(
	apiRetryFacade retryFacade.Interface,
	logger logger.Interface,
	symbol string,
) *Strategy {
	return &Strategy{
		APIRetryFacade: apiRetryFacade,
		Logger:         logger,
		Name:           MainStrategyName,
		Symbol:         symbol,
		Timeframe: types.Timeframe{
			Value: 1,
			Unit:  "m",
		},
	}
}
