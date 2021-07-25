package baseClass

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/logger"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"math"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// BaseClass ...
type BaseClass struct {
	APIRetryFacade          retryFacade.Interface
	API                     api.Interface
	Logger                  logger.Interface
	CandlesHandler          candlesHandler.Interface
	HorizontalLevelsService horizontalLevels.Interface
	TrendsService           trends.Interface

	Name      string
	Symbol    types.Symbol
	Timeframe types.Timeframe

	orders             []*api.Order
	currentBrokerQuote *api.Quote
	positions          []*api.Position
	state              *api.State
	currentPosition    *api.Position
	pendingOrder       *api.Order
	currentOrder       *api.Order
}

// SetCandlesHandler ...
func (s *BaseClass) SetCandlesHandler(candlesHandler candlesHandler.Interface) {
	s.CandlesHandler = candlesHandler
}

// SetHorizontalLevelsService ...
func (s *BaseClass) SetHorizontalLevelsService(horizontalLevelsService horizontalLevels.Interface) {
	s.HorizontalLevelsService = horizontalLevelsService
}

// SetTrendsService ...
func (s *BaseClass) SetTrendsService(trendsService trends.Interface) {
	s.TrendsService = trendsService
}

// GetTimeframe ...
func (s *BaseClass) GetTimeframe() *types.Timeframe {
	return &s.Timeframe
}

// GetSymbol ...
func (s *BaseClass) GetSymbol() *types.Symbol {
	return &s.Symbol
}

// Initialize ...
func (s *BaseClass) Initialize() {

}

// DailyReset ...
func (s *BaseClass) DailyReset() {

}

// SetOrders ...
func (s *BaseClass) SetOrders(orders []*api.Order) {
	s.orders = orders
}

// GetOrders ...
func (s *BaseClass) GetOrders() []*api.Order {
	return s.orders
}

// SetCurrentBrokerQuote ...
func (s *BaseClass) SetCurrentBrokerQuote(quote *api.Quote) {
	s.currentBrokerQuote = quote
}

// GetCurrentBrokerQuote ...
func (s *BaseClass) GetCurrentBrokerQuote() *api.Quote {
	return s.currentBrokerQuote
}

// SetPositions ...
func (s *BaseClass) SetPositions(positions []*api.Position) {
	s.positions = positions
}

// GetPositions ...
func (s *BaseClass) GetPositions() []*api.Position {
	return s.positions
}

// SetState ...
func (s *BaseClass) SetState(state *api.State) {
	s.state = state
}

// GetState ...
func (s *BaseClass) GetState() *api.State {
	return s.state
}

// OnReceiveMarketData ...
func (s *BaseClass) OnReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {

}

func (s *BaseClass) Log(strategyName string, message string, logType ...logger.LogType) {
	s.Logger.Log(strategyName+" - "+message, logType...)
}

func (s *BaseClass) SetStringValues(order *api.Order) {
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
		limitPrice := utils.FloatToString(math.Round(float64(*order.LimitPrice)*10)/10, symbol.PriceDecimals)
		order.StringValues.LimitPrice = &limitPrice
	} else {
		stopPrice := utils.FloatToString(math.Round(float64(*order.StopPrice)*10)/10, symbol.PriceDecimals)
		order.StringValues.StopPrice = &stopPrice
	}
	if order.StopLoss != nil {
		stopLossPrice := utils.FloatToString(math.Round(float64(*order.StopLoss)*10)/10, symbol.PriceDecimals)
		order.StringValues.StopLoss = &stopLossPrice
	}
	if order.TakeProfit != nil {
		takeProfitPrice := utils.FloatToString(math.Round(float64(*order.TakeProfit)*10)/10, symbol.PriceDecimals)
		order.StringValues.TakeProfit = &takeProfitPrice
	}
}

func (s *BaseClass) CheckIfSLShouldBeAdjusted(
	params *types.StrategyParams,
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

func (s *BaseClass) CheckNewestOpenedPositionSLandTP(longParams *types.StrategyParams, shortParams *types.StrategyParams) {
	for {
		position := utils.FindPositionBySymbol(s.GetPositions(), s.GetSymbol().BrokerAPIName)

		if position != nil && s.currentPosition == nil {
			s.currentPosition = position

			var tp string
			var sl string
			var closePosition bool = false

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

			if closePosition {
				s.Log(s.Name, "Will immediately close the position since it was executed very far away from the stop price")
				s.Log(s.Name, "Order is "+utils.GetStringRepresentation(s.currentOrder))
				s.Log(s.Name, "Position is "+utils.GetStringRepresentation(s.currentPosition))

				workingOrders := s.API.GetWorkingOrders(utils.FilterOrdersBySymbol(s.GetOrders(), s.GetSymbol().BrokerAPIName))
				s.Log(s.Name, "Closing working orders first ... "+utils.GetStringRepresentation(workingOrders))

				s.APIRetryFacade.CloseOrders(
					workingOrders,
					retryFacade.RetryParams{
						DelayBetweenRetries: 5 * time.Second,
						MaxRetries:          30,
						SuccessCallback: func() {
							s.SetOrders(nil)
							s.SetPendingOrder(nil)

							s.Log(s.Name, "Closed all orders. Closing the position now ... ")
							s.APIRetryFacade.ClosePosition(s.currentPosition.Instrument, retryFacade.RetryParams{
								DelayBetweenRetries: 5 * time.Second,
								MaxRetries:          20,
							})
						},
					})
			} else {
				s.Log(s.Name, "Modifying the SL and TP of the recently open position ... ")
				s.APIRetryFacade.ModifyPosition(s.GetSymbol().BrokerAPIName, tp, sl, retryFacade.RetryParams{
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

func (s *BaseClass) GetPendingOrder() *api.Order {
	return s.pendingOrder
}

func (s *BaseClass) SetPendingOrder(order *api.Order) {
	s.pendingOrder = order
}

func (s *BaseClass) GetCurrentOrder() *api.Order {
	return s.currentOrder
}

func (s *BaseClass) SetCurrentOrder(order *api.Order) {
	s.currentOrder = order
}
