package markets

import (
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/positionSize"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"math"
	"strconv"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

type BaseMarketClass struct {
	Container *services.Container

	CandlesHandler candlesHandler.Interface

	MarketData           types.MarketData
	ToExecuteOnNewCandle func()

	currentPosition           *api.Position
	currentPositionExecutedAt time.Time
	pendingOrder              *api.Order
	currentOrder              *api.Order

	isReady    bool
	lastVolume float64

	mutex *sync.Mutex
}

func (s *BaseMarketClass) SetContainer(c *services.Container) {
	s.Container = c
}
func (s *BaseMarketClass) GetContainer() *services.Container {
	return s.Container
}

func (s *BaseMarketClass) SetCandlesHandler(c candlesHandler.Interface) {
	s.CandlesHandler = c
}

func (s *BaseMarketClass) GetCandlesHandler() candlesHandler.Interface {
	return s.CandlesHandler
}

func (s *BaseMarketClass) GetMarketData() *types.MarketData {
	return &s.MarketData
}

func (s *BaseMarketClass) Initialize() {
	s.CandlesHandler.InitCandles(time.Now(), s.MarketData.CandlesFileName)
	go s.CheckNewestOpenedPositionSLandTP()

	s.mutex = &sync.Mutex{}
	s.isReady = true
}

func (s *BaseMarketClass) DailyReset() {
	var candlesMap = make(map[string][]int)

	// todo: refactor
	candlesMap["1m"] = []int{7200, 1441}
	candlesMap["3m"] = []int{5000, 481}
	candlesMap["5m"] = []int{3000, 289}
	candlesMap["15m"] = []int{1500, 97}
	candlesMap["30m"] = []int{1500, 49}
	candlesMap["1h"] = []int{1000, 25}
	candlesMap["2h"] = []int{1000, 13}
	candlesMap["4h"] = []int{1000, 7}
	candlesMap["8h"] = []int{500, 4}
	candlesMap["1d"] = []int{500, 2}

	key := strconv.Itoa(int(s.MarketData.Timeframe.Value)) + s.MarketData.Timeframe.Unit

	maxCandles := candlesMap[key][0]
	candlesToRemove := candlesMap[key][1]

	totalCandles := len(s.CandlesHandler.GetCompletedCandles())

	s.Log("Total candles is " + strconv.Itoa(totalCandles) + " - max candles is " + strconv.Itoa(maxCandles))
	if totalCandles < maxCandles {
		s.Log("Not removing any candles yet")
		return
	}

	s.Log("Removing old candles ... " + strconv.Itoa(int(candlesToRemove)))
	s.CandlesHandler.RemoveOldCandles(uint(candlesToRemove))
}

func (s *BaseMarketClass) SetCurrentPositionExecutedAt(timestamp int64) {
	s.currentPositionExecutedAt = time.Unix(timestamp, 0)
}

func (s *BaseMarketClass) OnReceiveMarketData(data *tradingviewsocket.QuoteData) {
	s.Log("Received data -> " + utils.GetStringRepresentation(data))

	if !s.isReady {
		s.Log("Not ready to process yet, doing nothing ...")
		return
	}

	s.mutex.Lock()
	defer func() {
		if data.Volume != nil {
			s.lastVolume = *data.Volume
		}

		s.Log("Candles amount -> " + strconv.Itoa(len(s.CandlesHandler.GetCompletedCandles())))
		s.mutex.Unlock()
	}()

	s.Log("Updating candles... ")
	s.CandlesHandler.UpdateCandles(data, s.lastVolume, s.OnNewCandle)
}

func (s *BaseMarketClass) OnNewCandle() {
	s.Log("\n\n")
	s.Log("New candle has been added " + time.Unix(s.CandlesHandler.GetLastCompletedCandle().Timestamp, 0).Format("02/01/2006 15:04:05"))

	s.Log("Last completed candles")
	candles := s.CandlesHandler.GetCompletedCandles()
	for i := len(candles) - 7; i < len(candles); i++ {
		if i >= 0 {
			s.Log(utils.GetStringRepresentation(candles[i]))
		}
	}
	s.Log("\n\n")

	s.ToExecuteOnNewCandle()
}

func (s *BaseMarketClass) Log(message string) {
	s.Container.Logger.Log(s.MarketData.SocketName+" - "+message, s.MarketData.LogType)
}

func (s *BaseMarketClass) SetStringValues(order *api.Order) {
	order.CurrentAsk = nil
	order.CurrentBid = nil

	currentAsk := "0.0"
	currentBid := "0.0"

	qty := utils.IntToString(int64(order.Qty))
	order.StringValues = &api.OrderStringValues{
		CurrentAsk: &currentAsk,
		CurrentBid: &currentBid,
		Qty:        &qty,
	}

	if s.Container.API.IsLimitOrder(order) {
		limitPrice := utils.FloatToString(*order.LimitPrice, s.MarketData.PriceDecimals)
		order.StringValues.LimitPrice = &limitPrice
	}
	if s.Container.API.IsStopOrder(order) {
		stopPrice := utils.FloatToString(*order.StopPrice, s.MarketData.PriceDecimals)
		order.StringValues.StopPrice = &stopPrice
	}
	if order.StopLoss != nil {
		stopLossPrice := utils.FloatToString(*order.StopLoss, s.MarketData.PriceDecimals)
		order.StringValues.StopLoss = &stopLossPrice
	}
	if order.TakeProfit != nil {
		takeProfitPrice := utils.FloatToString(*order.TakeProfit, s.MarketData.PriceDecimals)
		order.StringValues.TakeProfit = &takeProfitPrice
	}
}

func (s *BaseMarketClass) CheckNewestOpenedPositionSLandTP() {
	// TODO: MaxTradeExecutionPriceDifference.
	// TODO2: TakeProfit and StopLoss distance should be the correct ones.
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

func (s *BaseMarketClass) SavePendingOrder(side string, validTimes *types.TradingTimes) {
	go func() {
		s.Log("Save pending order called for side " + side)

		if utils.FindPositionByMarket(s.Container.APIData.GetPositions(), s.GetMarketData().BrokerAPIName) != nil {
			s.Log("Can't save pending order since there is an open position")
			return
		}

		workingOrders := s.Container.API.GetWorkingOrders(s.Container.APIData.GetOrders())

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

		slOrder, tpOrder := s.Container.API.GetSLAndTPOrders(mainOrder.ID, workingOrders)

		if slOrder != nil {
			mainOrder.StopLoss = slOrder.StopPrice
		}
		if tpOrder != nil {
			mainOrder.TakeProfit = tpOrder.LimitPrice
		}

		if s.Container.API.IsLimitOrder(mainOrder) {
			mainOrder.StopPrice = nil
		}
		if s.Container.API.IsStopOrder(mainOrder) {
			mainOrder.LimitPrice = nil
		}

		s.Container.APIRetryFacade.CloseOrders(
			s.Container.API.GetWorkingOrders(utils.FilterOrdersByMarket(s.Container.APIData.GetOrders(), s.GetMarketData().BrokerAPIName)),
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

	p := utils.FindPositionByMarket(s.Container.APIData.GetPositions(), s.GetMarketData().BrokerAPIName)
	if p != nil {
		s.Log("Can't create the pending order since there is an open position -> " + utils.GetStringRepresentation(p))
		return
	}

	go func(pendingOrder *api.Order) {
		var price float64
		if s.Container.API.IsStopOrder(pendingOrder) {
			price = *pendingOrder.StopPrice
		} else {
			price = *pendingOrder.LimitPrice
		}

		lastCompletedCandle := s.CandlesHandler.GetLastCompletedCandle()
		s.Log("Last completed candle -> " + utils.GetStringRepresentation(lastCompletedCandle))

		if side == ibroker.LongSide {
			if s.Container.API.IsStopOrder(pendingOrder) && price <= lastCompletedCandle.Close {
				s.Log("STOP ORDER -> Price is lower than last completed candle.close - Can't create the pending order")
				return
			}
			if s.Container.API.IsLimitOrder(pendingOrder) && price >= lastCompletedCandle.Close {
				s.Log("LIMIT ORDER -> Price is higher than last completed candle.close - Can't create the pending order")
				return
			}
		} else {
			if s.Container.API.IsStopOrder(pendingOrder) && price >= lastCompletedCandle.Close {
				s.Log("STOP ORDER -> Price is greater than last completed candle.close - Can't create the pending order")
				return
			}
			if s.Container.API.IsLimitOrder(pendingOrder) && price <= lastCompletedCandle.Close {
				s.Log("LIMIT ORDER -> Price is lower than last completed candle.close - Can't create the pending order")
				return
			}
		}

		s.Log("Everything is good - Creating the pending order")
		order := s.GetPendingOrder()
		s.Container.APIRetryFacade.CreateOrder(
			pendingOrder,
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

type OnValidTradeSetupParams struct {
	Price                float64
	StopLossDistance     float64
	TakeProfitDistance   float64
	RiskPercentage       float64
	IsValidTime          bool
	Side                 string
	WithPendingOrders    bool
	OrderType            string
	MinPositionSize      int64
	PositionSizeStrategy positionSize.Strategy
}

func (s *BaseMarketClass) OnValidTradeSetup(params OnValidTradeSetupParams) {
	var stopLoss float64
	var takeProfit float64

	if params.Side == ibroker.LongSide {
		stopLoss = params.Price - params.StopLossDistance
		takeProfit = params.Price + params.TakeProfitDistance
	} else {
		stopLoss = params.Price + params.StopLossDistance
		takeProfit = params.Price - params.TakeProfitDistance
	}

	size := s.Container.PositionSizeService.GetPositionSize(
		positionSize.GetPositionSizeParams{
			CurrentBalance:   s.Container.APIData.GetState().Equity,
			RiskPercentage:   params.RiskPercentage,
			StopLossDistance: params.StopLossDistance,
			MinPositionSize:  float64(params.MinPositionSize),
			EurExchangeRate:  s.MarketData.EurExchangeRate,
			Multiplier:       s.MarketData.PositionSizeMultiplier,
			Strategy:         params.PositionSizeStrategy,
		},
	)

	stopLoss = s.Round(stopLoss)
	takeProfit = s.Round(takeProfit)

	order := &api.Order{
		CurrentAsk: nil,
		CurrentBid: nil,
		Instrument: s.GetMarketData().BrokerAPIName,
		Qty:        size,
		Side:       params.Side,
		StopLoss:   &stopLoss,
		TakeProfit: &takeProfit,
		Type:       params.OrderType,
	}

	price := s.Round(params.Price)
	if s.Container.API.IsStopOrder(order) {
		order.StopPrice = &price
	}
	if s.Container.API.IsLimitOrder(order) {
		order.LimitPrice = &price
	}

	s.Log(params.Side + " order to be created -> " + utils.GetStringRepresentation(order))

	if params.WithPendingOrders {
		if utils.FindPositionByMarket(s.Container.APIData.GetPositions(), s.GetMarketData().BrokerAPIName) != nil {
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
	for _, p := range s.Container.APIData.GetPositions() {
		if p.Instrument == order.Instrument {
			position = p
		}
	}

	if position == nil {
		s.Log("There isn't any open position, let's create the order ...")

		s.Container.APIRetryFacade.CreateOrder(
			order,
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

	var diffInSeconds = s.CandlesHandler.GetLastCompletedCandle().Timestamp - s.currentPositionExecutedAt.Unix()
	s.Log("Difference in seconds is " + strconv.FormatInt(diffInSeconds, 10))

	if diffInSeconds > params.MaxSecondsOpenTrade {
		s.Log("Trade has been opened for too long, closing position ...")
		s.Container.APIRetryFacade.ClosePosition(position.Instrument, retryFacade.RetryParams{
			DelayBetweenRetries: 5 * time.Second,
			MaxRetries:          20,
		})
		s.Container.API.AddTrade(
			nil,
			position,
			func(price float64, order *api.Order) float64 {
				return price
			},
			s.MarketData.EurExchangeRate,
			s.CandlesHandler.GetLastCompletedCandle(),
			&s.MarketData,
		)
	} else {
		s.Log("Not closing the trade yet")
	}
}

func (s *BaseMarketClass) SetEmaCrossoverStrategyParams(longs *types.MarketStrategyParams, shorts *types.MarketStrategyParams) {
	s.MarketData.EmaCrossoverSetup.LongSetupParams = longs
	s.MarketData.EmaCrossoverSetup.ShortSetupParams = shorts
}
func (s *BaseMarketClass) SetRangesStrategyParams(longs *types.MarketStrategyParams, shorts *types.MarketStrategyParams) {
	s.MarketData.RangesSetup.LongSetupParams = longs
	s.MarketData.RangesSetup.ShortSetupParams = shorts
}

func (s *BaseMarketClass) Round(n float64) float64 {
	p := math.Pow(10, float64(s.MarketData.PriceDecimals))
	r := math.Round(n*p) / p
	return r
}
