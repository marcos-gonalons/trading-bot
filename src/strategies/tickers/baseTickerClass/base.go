package baseTickerClass

import (
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
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// BaseTickerClass ...
type BaseTickerClass struct {
	APIRetryFacade          retryFacade.Interface
	API                     api.Interface
	APIData                 api.DataInterface
	Logger                  logger.Interface
	CandlesHandler          candlesHandler.Interface
	HorizontalLevelsService horizontalLevels.Interface
	TrendsService           trends.Interface

	Name      string
	Symbol    types.Symbol
	Timeframe types.Timeframe

	currentExecutionTime      time.Time
	currentBrokerQuote        *api.Quote
	currentPosition           *api.Position
	currentPositionExecutedAt time.Time
	pendingOrder              *api.Order
	currentOrder              *api.Order

	eurExchangeRate float64
}

// SetCandlesHandler ...
func (s *BaseTickerClass) SetCandlesHandler(candlesHandler candlesHandler.Interface) {
	s.CandlesHandler = candlesHandler
}

// SetHorizontalLevelsService ...
func (s *BaseTickerClass) SetHorizontalLevelsService(horizontalLevelsService horizontalLevels.Interface) {
	s.HorizontalLevelsService = horizontalLevelsService
}

// SetTrendsService ...
func (s *BaseTickerClass) SetTrendsService(trendsService trends.Interface) {
	s.TrendsService = trendsService
}

// GetTimeframe ...
func (s *BaseTickerClass) GetTimeframe() *types.Timeframe {
	return &s.Timeframe
}

// GetSymbol ...
func (s *BaseTickerClass) GetSymbol() *types.Symbol {
	return &s.Symbol
}

// Initialize ...
func (s *BaseTickerClass) Initialize() {
	s.SetEurExchangeRate(1)
}

// DailyReset ...
func (s *BaseTickerClass) DailyReset() {

}

// SetCurrentBrokerQuote ...
func (s *BaseTickerClass) SetCurrentBrokerQuote(quote *api.Quote) {
	s.currentBrokerQuote = quote
}

// GetCurrentBrokerQuote ...
func (s *BaseTickerClass) GetCurrentBrokerQuote() *api.Quote {
	return s.currentBrokerQuote
}

// OnReceiveMarketData ...
func (s *BaseTickerClass) OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.Log("", "Received data -> "+utils.GetStringRepresentation(data))

	s.SetCurrentExecutionTime(time.Now())
}

func (s *BaseTickerClass) Log(strategyName string, message string) {
	s.Logger.Log(strategyName+" - "+message, s.GetSymbol().LogType)
}

func (s *BaseTickerClass) SetStringValues(order *api.Order) {
	symbol := s.GetSymbol()

	order.CurrentAsk = &s.currentBrokerQuote.Ask
	order.CurrentBid = &s.currentBrokerQuote.Bid

	currentAsk := utils.FloatToString(float64(*order.CurrentAsk), symbol.PriceDecimals)
	currentBid := utils.FloatToString(float64(*order.CurrentBid), symbol.PriceDecimals)
	qty := utils.IntToString(int64(order.Qty))
	order.StringValues = &api.OrderStringValues{
		CurrentAsk: &currentAsk,
		CurrentBid: &currentBid,
		Qty:        &qty,
	}

	if s.API.IsLimitOrder(order) {
		limitPrice := utils.FloatToString(float64(*order.LimitPrice), symbol.PriceDecimals)
		order.StringValues.LimitPrice = &limitPrice
	} else {
		stopPrice := utils.FloatToString(float64(*order.StopPrice), symbol.PriceDecimals)
		order.StringValues.StopPrice = &stopPrice
	}
	if order.StopLoss != nil {
		stopLossPrice := utils.FloatToString(float64(*order.StopLoss), symbol.PriceDecimals)
		order.StringValues.StopLoss = &stopLossPrice
	}
	if order.TakeProfit != nil {
		takeProfitPrice := utils.FloatToString(float64(*order.TakeProfit), symbol.PriceDecimals)
		order.StringValues.TakeProfit = &takeProfitPrice
	}
}

func (s *BaseTickerClass) CheckIfSLShouldBeAdjusted(
	params *types.TickerStrategyParams,
	position *api.Position,
) {
	if params.TPDistanceShortForTighterSL <= 0 {
		return
	}

	s.Log(s.Name, "Checking if the position needs to have the SL adjusted with this params ... "+utils.GetStringRepresentation(params))
	s.Log(s.Name, "Position is "+utils.GetStringRepresentation(position))

	_, tpOrder := s.API.GetBracketOrdersForOpenedPosition(position)

	if tpOrder == nil {
		s.Log(s.Name, "Take Profit order not found ...")
		return
	}

	shouldBeAdjusted := false
	if s.API.IsLongPosition(position) {
		shouldBeAdjusted = float64(*tpOrder.LimitPrice)-s.CandlesHandler.GetLastCandle().High < params.TPDistanceShortForTighterSL
	} else {
		shouldBeAdjusted = s.CandlesHandler.GetLastCandle().Low-float64(*tpOrder.LimitPrice) < params.TPDistanceShortForTighterSL
	}

	if shouldBeAdjusted {
		s.Log(s.Name, "The price is very close to the TP. Adjusting SL...")

		s.APIRetryFacade.ModifyPosition(
			s.GetSymbol().BrokerAPIName,
			utils.FloatToString(float64(*tpOrder.LimitPrice), 2),
			utils.FloatToString(float64(position.AvgPrice)+params.SLDistanceWhenTPIsVeryClose, 2),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          20,
			},
		)
	} else {
		s.Log(s.Name, "The price is not close to the TP yet. Doing nothing ...")
	}
}

func (s *BaseTickerClass) CheckNewestOpenedPositionSLandTP(longParams *types.TickerStrategyParams, shortParams *types.TickerStrategyParams) {
	for {
		s.Log(s.Name, "Checking newest open position")
		position := utils.FindPositionBySymbol(s.APIData.GetPositions(), s.GetSymbol().BrokerAPIName)
		s.Log(s.Name, "Position ->"+utils.GetStringRepresentation(position))
		s.Log(s.Name, "Current position ->"+utils.GetStringRepresentation(s.currentPosition))

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
					tp = utils.FloatToString(float64(s.currentPosition.AvgPrice-shortParams.TakeProfitDistance), s.GetSymbol().PriceDecimals)
					sl = utils.FloatToString(float64(s.currentPosition.AvgPrice+shortParams.StopLossDistance), s.GetSymbol().PriceDecimals)
				} else {
					if s.API.IsStopOrder(s.currentOrder) && float64(s.currentPosition.AvgPrice-*s.currentOrder.StopPrice) > longParams.MaxTradeExecutionPriceDifference {
						closePosition = true
					}
					tp = utils.FloatToString(float64(s.currentPosition.AvgPrice+longParams.TakeProfitDistance), s.GetSymbol().PriceDecimals)
					sl = utils.FloatToString(float64(s.currentPosition.AvgPrice-longParams.StopLossDistance), s.GetSymbol().PriceDecimals)
				}
			} else {
				s.Log(s.Name, "current order is nil (because the order was created manually on the broker)")
			}

			if closePosition {
				s.Log(s.Name, "Will immediately close the position since it was executed very far away from the stop price")
				s.Log(s.Name, "Order is "+utils.GetStringRepresentation(s.currentOrder))
				s.Log(s.Name, "Position is "+utils.GetStringRepresentation(s.currentPosition))

				workingOrders := s.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.APIData.GetOrders(), s.GetSymbol().BrokerAPIName))
				s.Log(s.Name, "Closing working orders first ... "+utils.GetStringRepresentation(workingOrders))

				s.APIRetryFacade.CloseOrders(
					workingOrders,
					retryFacade.RetryParams{
						DelayBetweenRetries: 5 * time.Second,
						MaxRetries:          30,
						SuccessCallback: func() {
							s.SetPendingOrder(nil)

							s.Log(s.Name, "Closed all orders. Closing the position now ... ")
							s.APIRetryFacade.ClosePosition(s.currentPosition.Instrument, retryFacade.RetryParams{
								DelayBetweenRetries: 5 * time.Second,
								MaxRetries:          20,
							})
						},
					})
			} else {
				if s.currentOrder != nil {
					s.Log(s.Name, "Modifying the SL and TP of the recently opened position ... ")
					s.APIRetryFacade.ModifyPosition(s.GetSymbol().BrokerAPIName, tp, sl, retryFacade.RetryParams{
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

func (s *BaseTickerClass) GetPendingOrder() *api.Order {
	return s.pendingOrder
}

func (s *BaseTickerClass) SetPendingOrder(order *api.Order) {
	s.pendingOrder = order
}

func (s *BaseTickerClass) GetCurrentOrder() *api.Order {
	return s.currentOrder
}

func (s *BaseTickerClass) SetCurrentOrder(order *api.Order) {
	s.currentOrder = order
}

func (s *BaseTickerClass) SetCurrentExecutionTime(t time.Time) {
	s.currentExecutionTime = t
}

func (s *BaseTickerClass) GetCurrentExecutionTime() time.Time {
	return s.currentExecutionTime
}

func (s *BaseTickerClass) SetEurExchangeRate(rate float64) {
	s.eurExchangeRate = rate
}

func (s *BaseTickerClass) GetEurExchangeRate() float64 {
	return s.eurExchangeRate
}

func (s *BaseTickerClass) SavePendingOrder(side string, validTimes types.TradingTimes) {
	go func() {
		s.Log(s.Name, "Save pending order called for side "+side)

		if utils.FindPositionBySymbol(s.APIData.GetPositions(), s.GetSymbol().BrokerAPIName) != nil {
			s.Log(s.Name, "Can't save pending order since there is an open position")
			return
		}

		workingOrders := s.API.GetWorkingOrders(s.APIData.GetOrders())

		if len(workingOrders) == 0 {
			s.Log(s.Name, "There aren't any working orders, doing nothing ...")
			return
		}

		var mainOrder *api.Order
		for _, workingOrder := range workingOrders {
			if workingOrder.Side == side && workingOrder.ParentID == nil {
				mainOrder = workingOrder
			}
		}

		if mainOrder == nil {
			s.Log(s.Name, "There isn't an active order for this side "+side)
			return
		}

		if utils.IsExecutionTimeValid(
			s.currentExecutionTime,
			[]string{},
			[]string{},
			validTimes.ValidHalfHours,
		) {
			s.Log(s.Name, "No need to save the pending order since we are in the right time")
			return
		}

		s.Log(s.Name, "Closing the current order and saving it for the future, since now it's not the time for profitable trading.")
		s.Log(s.Name, "This is the current order -> "+utils.GetStringRepresentation(mainOrder))

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
			s.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.APIData.GetOrders(), s.GetSymbol().BrokerAPIName)),
			retryFacade.RetryParams{
				DelayBetweenRetries: 5 * time.Second,
				MaxRetries:          30,
				SuccessCallback: func() {
					s.SetPendingOrder(mainOrder)
					s.Log(s.Name, "Closed all working orders correctly and pending order saved -> "+utils.GetStringRepresentation(s.GetPendingOrder()))
				},
			},
		)
	}()
}

func (s *BaseTickerClass) CreatePendingOrder(side string) {
	if s.GetPendingOrder().Side != side {
		return
	}

	p := utils.FindPositionBySymbol(s.APIData.GetPositions(), s.GetSymbol().BrokerAPIName)
	if p != nil {
		s.Log(s.Name, "Can't create the pending order since there is an open position -> "+utils.GetStringRepresentation(p))
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
		s.Log(s.Name, "Last completed candle -> "+utils.GetStringRepresentation(lastCompletedCandle))

		if side == ibroker.LongSide {
			if s.API.IsStopOrder(pendingOrder) && price <= float32(lastCompletedCandle.Close) {
				s.Log(s.Name, "STOP ORDER -> Price is lower than last completed candle.close - Can't create the pending order")
				return
			}
			if s.API.IsLimitOrder(pendingOrder) && price >= float32(lastCompletedCandle.Close) {
				s.Log(s.Name, "LIMIT ORDER -> Price is higher than last completed candle.close - Can't create the pending order")
				return
			}
		} else {
			if s.API.IsStopOrder(pendingOrder) && price >= float32(lastCompletedCandle.Close) {
				s.Log(s.Name, "STOP ORDER -> Price is greater than last completed candle.close - Can't create the pending order")
				return
			}
			if s.API.IsLimitOrder(pendingOrder) && price <= float32(lastCompletedCandle.Close) {
				s.Log(s.Name, "LIMIT ORDER -> Price is lower than last completed candle.close - Can't create the pending order")
				return
			}
		}

		s.Log(s.Name, "Everything is good - Creating the pending order")
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
						s.Log(s.Name, "Pending order successfully created ... "+utils.GetStringRepresentation(s.GetCurrentOrder()))
					}
				}(order),
			},
		)
	}(s.GetPendingOrder())

	s.SetPendingOrder(nil)
}

type OnValidTradeSetupParams struct {
	Price              float64
	StrategyName       string
	StopLossDistance   float32
	TakeProfitDistance float32
	RiskPercentage     float64
	IsValidTime        bool
	Side               string
	WithPendingOrders  bool
	OrderType          string
	MinPositionSize    int64
}

func (s *BaseTickerClass) OnValidTradeSetup(params OnValidTradeSetupParams) {
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
		Instrument: s.GetSymbol().BrokerAPIName,
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

	s.Log(params.StrategyName, params.Side+" order to be created -> "+utils.GetStringRepresentation(order))

	if params.WithPendingOrders {
		if utils.FindPositionBySymbol(s.APIData.GetPositions(), s.GetSymbol().BrokerAPIName) != nil {
			s.Log(params.StrategyName, "There is an open position, saving the order for later ...")
			s.SetPendingOrder(order)
			return
		}

		if !params.IsValidTime {
			s.Log(params.StrategyName, "Now is not the time for opening any "+params.Side+" orders, saving it for later ...")
			s.SetPendingOrder(order)
			return
		}
	}

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
					s.Log(params.StrategyName, "New order successfully created ... "+utils.GetStringRepresentation(s.GetCurrentOrder()))
				}
			}(order),
		},
	)
}

func (s *BaseTickerClass) CheckOpenPositionTTL(params *types.TickerStrategyParams, position *api.Position) {
	if params.MaxSecondsOpenTrade == 0 {
		return
	}
	if s.currentPosition == nil {
		s.Log(s.Name, "CheckOpenPositionTTL called, but currentPosition is nil")
		return
	}

	s.Log(s.Name, "Checking open position TTL, it was opened on "+s.currentPositionExecutedAt.Format("2006-01-02 15:04:05"))
	s.Log(s.Name, "Position is "+utils.GetStringRepresentation(position))
	s.Log(s.Name, "Max seconds open trade is"+strconv.FormatInt(params.MaxSecondsOpenTrade, 10))

	var diffInSeconds = time.Now().Unix() - s.currentPositionExecutedAt.Unix()
	s.Log(s.Name, "Difference in seconds is "+strconv.FormatInt(diffInSeconds, 10))

	if diffInSeconds > params.MaxSecondsOpenTrade {
		s.Log(s.Name, "Trade has been opened for too long, closing position ...")
		s.APIRetryFacade.ClosePosition(s.currentPosition.Instrument, retryFacade.RetryParams{
			DelayBetweenRetries: 5 * time.Second,
			MaxRetries:          20,
		})
	} else {
		s.Log(s.Name, "Not closing the trade yet")
	}
}

func (s *BaseTickerClass) getSlAndTpOrders(
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
