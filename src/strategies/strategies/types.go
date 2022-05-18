package strategies

import (
	"TradingBot/src/strategies/markets/baseMarketClass"
	"TradingBot/src/types"
)

type StrategyParams struct {
	BaseMarketClass       *baseMarketClass.BaseMarketClass
	MarketStrategyParams  *types.MarketStrategyParams
	WithPendingOrders     bool
	CloseOrdersOnBadTrend bool
}
