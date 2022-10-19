package baseMarketClass

import (
	"TradingBot/src/markets/interfaces"
	"TradingBot/src/services/api"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/logger"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"strconv"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// BaseMarketClass ...
type BaseMarketClass struct {
	APIRetryFacade          retryFacade.Interface
	API                     api.Interface
	APIData                 api.DataInterface
	Logger                  logger.Interface
	CandlesHandler          candlesHandler.Interface
	HorizontalLevelsService horizontalLevels.Interface
	TrendsService           trends.Interface

	MarketData types.MarketData

	currentBrokerQuote        *api.Quote
	currentPosition           *api.Position
	currentPositionExecutedAt time.Time
	pendingOrder              *api.Order
	currentOrder              *api.Order

	isReady           bool
	lastCandlesAmount int
	lastVolume        float64
	lastBid           *float64
	lastAsk           *float64

	mutex *sync.Mutex
}

// SetDependencies ...
func (s *BaseMarketClass) SetDependencies(d interfaces.MarketInstanceDependencies) {
	s.APIRetryFacade = d.APIRetryFacade
	s.API = d.API
	s.APIData = d.APIData
	s.Logger = d.Logger
	s.CandlesHandler = d.CandlesHandler
	s.HorizontalLevelsService = d.HorizontalLevelsService
	s.TrendsService = d.TrendsService
}

// GetCandlesHandler ...
func (s *BaseMarketClass) GetCandlesHandler() candlesHandler.Interface {
	return s.CandlesHandler
}

// GetMarketData ...
func (s *BaseMarketClass) GetMarketData() *types.MarketData {
	return &s.MarketData
}

// GetAPIData ...
func (s *BaseMarketClass) GetAPIData() api.DataInterface {
	return s.APIData
}

// Initialize ...
func (s *BaseMarketClass) Initialize() {
	s.CandlesHandler.InitCandles(time.Now(), s.MarketData.CandlesFileName)
	go s.CheckNewestOpenedPositionSLandTP()

	s.mutex = &sync.Mutex{}
	s.isReady = true
}

// DailyReset ...
func (s *BaseMarketClass) DailyReset() {
	// todo:
	// must remove candles based on the marketdata.timeframe
	// following code removes candles for 1h timeframe
	// adapt it for any timeframe

	minCandles := 7 * 2 * 24
	totalCandles := len(s.CandlesHandler.GetCandles())

	s.Log("Total candles is " + strconv.Itoa(totalCandles) + " - min candles is " + strconv.Itoa(minCandles))
	if totalCandles < minCandles {
		s.Log("Not removing any candles yet")
		return
	}

	var candlesToRemove uint = 25
	s.Log("Removing old candles ... " + strconv.Itoa(int(candlesToRemove)))
	s.CandlesHandler.RemoveOldCandles(candlesToRemove)
}

// SetCurrentPositionExecutedAt ...
func (s *BaseMarketClass) SetCurrentPositionExecutedAt(timestamp int64) {
	s.currentPositionExecutedAt = time.Unix(timestamp, 0)
}

// SetCurrentBrokerQuote ...
func (s *BaseMarketClass) SetCurrentBrokerQuote(quote *api.Quote) {
	s.currentBrokerQuote = quote
}

// GetCurrentBrokerQuote ...
func (s *BaseMarketClass) GetCurrentBrokerQuote() *api.Quote {
	return s.currentBrokerQuote
}

// OnReceiveMarketData ...
func (s *BaseMarketClass) OnReceiveMarketData(data *tradingviewsocket.QuoteData) {
	s.Log("Received data -> " + utils.GetStringRepresentation(data))

	if !s.isReady {
		s.Log("Not ready to process yet, doing nothing ...")
		return
	}

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

		s.lastCandlesAmount = len(s.CandlesHandler.GetCandles())
		s.Log("Candles amount -> " + strconv.Itoa(s.lastCandlesAmount))
	}()

	s.Log("Updating candles... ")
	s.CandlesHandler.UpdateCandles(data, s.lastVolume)

	if s.lastCandlesAmount != len(s.CandlesHandler.GetCandles()) {
		s.OnNewCandle()
	} else {
		s.Log("Doing nothing - still same candle")
	}
}

// OnNewCandle ...
func (s *BaseMarketClass) OnNewCandle() {
	s.Log("New candle has been added. Executing strategy code ...")
}

func (s *BaseMarketClass) Log(message string) {
	s.Logger.Log(s.MarketData.SocketName+" - "+message, s.GetMarketData().LogType)
}

func (s *BaseMarketClass) SetStringValues(order *api.Order) {
	market := s.GetMarketData()

	order.CurrentAsk = &s.currentBrokerQuote.Ask
	order.CurrentBid = &s.currentBrokerQuote.Bid

	currentAsk := utils.FloatToString(float64(*order.CurrentAsk), market.PriceDecimals)
	currentBid := utils.FloatToString(float64(*order.CurrentBid), market.PriceDecimals)
	qty := utils.IntToString(int64(order.Qty))
	order.StringValues = &api.OrderStringValues{
		CurrentAsk: &currentAsk,
		CurrentBid: &currentBid,
		Qty:        &qty,
	}

	if s.API.IsLimitOrder(order) {
		limitPrice := utils.FloatToString(float64(*order.LimitPrice), market.PriceDecimals)
		order.StringValues.LimitPrice = &limitPrice
	} else {
		stopPrice := utils.FloatToString(float64(*order.StopPrice), market.PriceDecimals)
		order.StringValues.StopPrice = &stopPrice
	}
	if order.StopLoss != nil {
		stopLossPrice := utils.FloatToString(float64(*order.StopLoss), market.PriceDecimals)
		order.StringValues.StopLoss = &stopLossPrice
	}
	if order.TakeProfit != nil {
		takeProfitPrice := utils.FloatToString(float64(*order.TakeProfit), market.PriceDecimals)
		order.StringValues.TakeProfit = &takeProfitPrice
	}
}

func (s *BaseMarketClass) CheckIfSLShouldBeAdjusted(
	params *types.MarketStrategyParams,
	position *api.Position,
) {
	if params.TrailingStopLoss != nil && params.TrailingStopLoss.TPDistanceShortForTighterSL <= 0 {
		return
	}

	s.Log("Checking if the position needs to have the SL adjusted with this params ... " + utils.GetStringRepresentation(params))
	s.Log("Position is " + utils.GetStringRepresentation(position))

	_, tpOrder := s.API.GetBracketOrdersForOpenedPosition(position)

	if tpOrder == nil {
		s.Log("Take Profit order not found ...")
		return
	}

	shouldBeAdjusted := false
	if s.API.IsLongPosition(position) {
		shouldBeAdjusted = float64(*tpOrder.LimitPrice)-s.CandlesHandler.GetLastCandle().High < params.TrailingStopLoss.TPDistanceShortForTighterSL
	} else {
		shouldBeAdjusted = s.CandlesHandler.GetLastCandle().Low-float64(*tpOrder.LimitPrice) < params.TrailingStopLoss.TPDistanceShortForTighterSL
	}

	if shouldBeAdjusted {
		s.Log("The price is very close to the TP. Adjusting SL...")

		s.APIRetryFacade.ModifyPosition(
			s.GetMarketData().BrokerAPIName,
			utils.FloatToString(float64(*tpOrder.LimitPrice), s.GetMarketData().PriceDecimals),
			utils.FloatToString(float64(position.AvgPrice)+params.TrailingStopLoss.SLDistanceWhenTPIsVeryClose, s.GetMarketData().PriceDecimals),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			},
		)
	} else {
		s.Log("The price is not close to the TP yet. Doing nothing ...")
	}
}

func (s *BaseMarketClass) CheckNewestOpenedPositionSLandTP() {
	longParams := s.MarketData.LongSetupParams
	shortParams := s.MarketData.ShortSetupParams

	for {
		positions := s.APIData.GetPositions()
		marketName := s.GetMarketData().BrokerAPIName
		s.Log("Checking newest open position")
		s.Log("Positions is -> " + utils.GetStringRepresentation(positions))
		s.Log("Market name is -> " + marketName)
		position := utils.FindPositionByMarket(positions, marketName)
		s.Log("Position ->" + utils.GetStringRepresentation(position))
		s.Log("Current position ->" + utils.GetStringRepresentation(s.currentPosition))

		if position != nil && s.currentPosition == nil {
			s.currentPosition = position
			s.currentPositionExecutedAt = time.Now()

			var tp string
			var sl string
			var closePosition bool = false

			if s.currentOrder != nil {
				if s.API.IsShortPosition(s.currentPosition) {
					if s.API.IsStopOrder(s.currentOrder) && float64(*s.currentOrder.StopPrice-s.currentPosition.AvgPrice) > shortParams.MaxTradeExecutionPriceDifference {
						closePosition = true
					}
					tp = utils.FloatToString(float64(s.currentPosition.AvgPrice-shortParams.TakeProfitDistance), s.GetMarketData().PriceDecimals)
					sl = utils.FloatToString(float64(s.currentPosition.AvgPrice+shortParams.StopLossDistance), s.GetMarketData().PriceDecimals)
				} else {
					if s.API.IsStopOrder(s.currentOrder) && float64(s.currentPosition.AvgPrice-*s.currentOrder.StopPrice) > longParams.MaxTradeExecutionPriceDifference {
						closePosition = true
					}
					tp = utils.FloatToString(float64(s.currentPosition.AvgPrice+longParams.TakeProfitDistance), s.GetMarketData().PriceDecimals)
					sl = utils.FloatToString(float64(s.currentPosition.AvgPrice-longParams.StopLossDistance), s.GetMarketData().PriceDecimals)
				}
			} else {
				s.Log("current order is nil (because the order was created manually on the broker)")
			}

			if closePosition {
				s.Log("Will immediately close the position since it was executed very far away from the stop price")
				s.Log("Order is " + utils.GetStringRepresentation(s.currentOrder))
				s.Log("Position is " + utils.GetStringRepresentation(s.currentPosition))

				workingOrders := s.API.GetWorkingOrders(utils.FilterOrdersByMarket(s.APIData.GetOrders(), s.GetMarketData().BrokerAPIName))
				s.Log("Closing working orders first ... " + utils.GetStringRepresentation(workingOrders))

				s.APIRetryFacade.CloseOrders(
					workingOrders,
					retryFacade.RetryParams{
						DelayBetweenRetries: 5 * time.Second,
						MaxRetries:          30,
						SuccessCallback: func() {
							s.SetPendingOrder(nil)

							s.Log("Closed all orders. Closing the position now ... ")
							s.APIRetryFacade.ClosePosition(s.currentPosition.Instrument, retryFacade.RetryParams{
								DelayBetweenRetries: 5 * time.Second,
								MaxRetries:          20,
							})
							s.API.AddTrade(
								nil,
								s.currentPosition,
								func(price float32, order *api.Order) float32 {
									return price
								},
								s.GetEurExchangeRate(),
								s.CandlesHandler.GetLastCandle(),
								s.GetMarketData(),
							)
						},
					})
			} else {
				if s.currentOrder != nil {
					// TODO: investigate why this executed at 23:00, causing an error saying
					// that it can't be traded at 23:00 (since it's not market hours),
					// causing the app to panic after reaching lot's of unsuccessful tries
					s.Log("Modifying the SL and TP of the recently opened position ... ")
					s.APIRetryFacade.ModifyPosition(s.GetMarketData().BrokerAPIName, tp, sl, retryFacade.RetryParams{
						DelayBetweenRetries: 5 * time.Second,
						MaxRetries:          20,
					})
				}
			}
		}

		if position == nil {
			s.currentPosition = nil
		}

		time.Sleep(5 * time.Second)
	}
}

func (s *BaseMarketClass) GetPendingOrder() *api.Order {
	return s.pendingOrder
}

func (s *BaseMarketClass) SetPendingOrder(order *api.Order) {
	s.pendingOrder = order
}

func (s *BaseMarketClass) GetCurrentOrder() *api.Order {
	return s.currentOrder
}

func (s *BaseMarketClass) SetCurrentOrder(order *api.Order) {
	s.currentOrder = order
}

func (s *BaseMarketClass) GetEurExchangeRate() float64 {
	return s.MarketData.EurExchangeRate
}

func (s *BaseMarketClass) SavePendingOrder(side string, validTimes *types.TradingTimes) {
	go func() {
		s.Log("Save pending order called for side " + side)

		if utils.FindPositionByMarket(s.APIData.GetPositions(), s.GetMarketData().BrokerAPIName) != nil {
			s.Log("Can't save pending order since there is an open position")
			return
		}

		workingOrders := s.API.GetWorkingOrders(s.APIData.GetOrders())

		if len(workingOrders) == 0 {
			s.Log("There aren't any working orders, doing nothing ...")
			return
		}

		var mainOrder *api.Order
		for _, workingOrder := range workingOrders {
			if workingOrder.Side == side && workingOrder.ParentID == nil {
				mainOrder = workingOrder
			}
		}

		if mainOrder == nil {
			s.Log("There isn't an active order for this side " + side)
			return
		}

		validHalfHours := []string{}
		if validTimes != nil {
			validHalfHours = validTimes.ValidHalfHours
		}
		if utils.IsExecutionTimeValid(
			time.Now(),
			[]string{},
			[]string{},
			validHalfHours,
		) {
			s.Log("No need to save the pending order since we are in the right time")
			return
		}

		s.Log("Closing the current order and saving it for the future, since now it's not the time for profitable trading.")
		s.Log("This is the current order -> " + utils.GetStringRepresentation(mainOrder))

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
			s.API.GetWorkingOrders(utils.FilterOrdersByMarket(s.APIData.GetOrders(), s.GetMarketData().BrokerAPIName)),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					s.SetPendingOrder(mainOrder)
					s.Log("Closed all working orders correctly and pending order saved -> " + utils.GetStringRepresentation(s.GetPendingOrder()))
				},
			},
		)
	}()
}

func (s *BaseMarketClass) CreatePendingOrder(side string) {
	if s.GetPendingOrder().Side != side {
		return
	}

	p := utils.FindPositionByMarket(s.APIData.GetPositions(), s.GetMarketData().BrokerAPIName)
	if p != nil {
		s.Log("Can't create the pending order since there is an open position -> " + utils.GetStringRepresentation(p))
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
		s.Log("Last completed candle -> " + utils.GetStringRepresentation(lastCompletedCandle))

		if side == ibroker.LongSide {
			if s.API.IsStopOrder(pendingOrder) && price <= float32(lastCompletedCandle.Close) {
				s.Log("STOP ORDER -> Price is lower than last completed candle.close - Can't create the pending order")
				return
			}
			if s.API.IsLimitOrder(pendingOrder) && price >= float32(lastCompletedCandle.Close) {
				s.Log("LIMIT ORDER -> Price is higher than last completed candle.close - Can't create the pending order")
				return
			}
		} else {
			if s.API.IsStopOrder(pendingOrder) && price >= float32(lastCompletedCandle.Close) {
				s.Log("STOP ORDER -> Price is greater than last completed candle.close - Can't create the pending order")
				return
			}
			if s.API.IsLimitOrder(pendingOrder) && price <= float32(lastCompletedCandle.Close) {
				s.Log("LIMIT ORDER -> Price is lower than last completed candle.close - Can't create the pending order")
				return
			}
		}

		s.Log("Everything is good - Creating the pending order")
		order := s.GetPendingOrder()
		s.APIRetryFacade.CreateOrder(
			pendingOrder,
			func() *api.Quote {
				return s.GetCurrentBrokerQuote()
			},
			s.SetStringValues,
			retryFacade.RetryParams{
				DelayBetweenRetries: 10 * time.Second,
				MaxRetries:          20,
				SuccessCallback: func(order *api.Order) func() {
					return func() {
						s.SetCurrentOrder(order)
						s.Log("Pending order successfully created ... " + utils.GetStringRepresentation(s.GetCurrentOrder()))
					}
				}(order),
			},
		)
	}(s.GetPendingOrder())

	s.SetPendingOrder(nil)
}

func (s *BaseMarketClass) OnValidTradeSetup(params interfaces.OnValidTradeSetupParams) {
	float32Price := float32(params.Price)

	var stopLoss float32
	var takeProfit float32

	if params.Side == ibroker.LongSide {
		stopLoss = float32Price - float32(params.StopLossDistance)
		takeProfit = float32Price + float32(params.TakeProfitDistance)
	} else {
		stopLoss = float32Price + float32(params.StopLossDistance)
		takeProfit = float32Price - float32(params.TakeProfitDistance)
	}

	size := utils.GetPositionSize(
		s.APIData.GetState().Equity,
		params.RiskPercentage,
		float64(params.StopLossDistance),
		float64(params.MinPositionSize),
		s.GetEurExchangeRate(),
	)

	order := &api.Order{
		CurrentAsk: &s.GetCurrentBrokerQuote().Ask,
		CurrentBid: &s.GetCurrentBrokerQuote().Bid,
		Instrument: s.GetMarketData().BrokerAPIName,
		Qty:        size,
		Side:       params.Side,
		StopLoss:   &stopLoss,
		TakeProfit: &takeProfit,
		Type:       params.OrderType,
	}
	if s.API.IsStopOrder(order) {
		order.StopPrice = &float32Price
	}
	if s.API.IsLimitOrder(order) {
		order.LimitPrice = &float32Price
	}

	s.Log(params.Side + " order to be created -> " + utils.GetStringRepresentation(order))

	if params.WithPendingOrders {
		if utils.FindPositionByMarket(s.APIData.GetPositions(), s.GetMarketData().BrokerAPIName) != nil {
			s.Log("There is an open position, saving the order for later ...")
			s.SetPendingOrder(order)
			return
		}

		if !params.IsValidTime {
			s.Log("Now is not the time for opening any " + params.Side + " orders, saving it for later ...")
			s.SetPendingOrder(order)
			return
		}
	}

	var position *api.Position
	for _, p := range s.APIData.GetPositions() {
		if p.Instrument == order.Instrument {
			position = p
		}
	}

	if position == nil {
		s.Log("There isn't any open position, let's create the order ...")

		s.APIRetryFacade.CreateOrder(
			order,
			func() *api.Quote {
				return s.GetCurrentBrokerQuote()
			},
			s.SetStringValues,
			retryFacade.RetryParams{
				DelayBetweenRetries: 10 * time.Second,
				MaxRetries:          20,
				SuccessCallback: func(order *api.Order) func() {
					return func() {
						s.SetCurrentOrder(order)
						s.Log("New order successfully created ... " + utils.GetStringRepresentation(s.GetCurrentOrder()))
					}
				}(order),
			},
		)
	} else {
		s.Log("Not creating the order since there is an open position")
		s.Log("Position is -> " + utils.GetStringRepresentation(s.currentPosition))
	}

}

func (s *BaseMarketClass) CheckOpenPositionTTL(params *types.MarketStrategyParams, position *api.Position) {
	if params.MaxSecondsOpenTrade == 0 {
		return
	}

	s.Log("Checking open position TTL, it was opened on " + s.currentPositionExecutedAt.Format("2006-01-02 15:04:05"))
	s.Log("Position is " + utils.GetStringRepresentation(position))
	s.Log("Max seconds open trade is" + strconv.FormatInt(params.MaxSecondsOpenTrade, 10))

	var diffInSeconds = s.CandlesHandler.GetLastCandle().Timestamp - s.currentPositionExecutedAt.Unix()
	s.Log("Difference in seconds is " + strconv.FormatInt(diffInSeconds, 10))

	if diffInSeconds > params.MaxSecondsOpenTrade {
		s.Log("Trade has been opened for too long, closing position ...")
		s.APIRetryFacade.ClosePosition(position.Instrument, retryFacade.RetryParams{
			DelayBetweenRetries: 5 * time.Second,
			MaxRetries:          20,
		})
		candles := s.CandlesHandler.GetCandles()
		s.API.AddTrade(
			nil,
			position,
			func(price float32, order *api.Order) float32 {
				return price
			},
			s.GetEurExchangeRate(),
			candles[len(candles)-3],
			s.GetMarketData(),
		)
	} else {
		s.Log("Not closing the trade yet")
	}
}

// todo: move away from this base class, maybe utils or maybe API static method
func (s *BaseMarketClass) getSlAndTpOrders(
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
