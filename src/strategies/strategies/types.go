package strategies

import (
	"TradingBot/src/strategies/tickers/baseTickerClass"
	"TradingBot/src/types"
)

type StrategyParams struct {
	BaseTickerClass       baseTickerClass.BaseTickerClass
	TickerStrategyParams  *types.TickerStrategyParams
	WithPendingOrders     bool
	CloseOrdersOnBadTrend bool
}
