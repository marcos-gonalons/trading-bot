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

func (s *BaseClass) checkIfSLShouldBeMovedToBreakEven(
	distanceToTp float64,
	position *api.Position,
) {
	if distanceToTp <= 0 {
		return
	}

	s.Log(s.Name, "Checking if the current position needs to have the SL adjusted... ")
	s.Log(s.Name, "Current position is "+utils.GetStringRepresentation(position))

	_, tpOrder := s.API.GetBracketOrdersForOpenedPosition(position)

	if tpOrder == nil {
		s.Log(s.Name, "Take Profit order not found ...")
		return
	}

	shouldBeAdjusted := false
	if s.API.IsLongPosition(position) {
		shouldBeAdjusted = float64(*tpOrder.LimitPrice)-s.CandlesHandler.GetLastCandle().High < distanceToTp
	} else {
		shouldBeAdjusted = s.CandlesHandler.GetLastCandle().Low-float64(*tpOrder.LimitPrice) < distanceToTp
	}

	if shouldBeAdjusted {
		s.Log(s.Name, "The price is very close to the TP. Adjusting SL to break even ...")

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
		s.Log(s.Name, "The price is not close to the TP yet. Doing nothing ...")
	}
}

func (s *BaseClass) Log(strategyName string, message string) {
	s.Logger.Log(strategyName + " - " + message)
}

func (s *BaseClass) SetStringValues(order *api.Order) {
	order.CurrentAsk = &s.currentBrokerQuote.Ask
	order.CurrentBid = &s.currentBrokerQuote.Bid

	utils.SetStringValues(order, s.GetSymbol(), s.API)
}
