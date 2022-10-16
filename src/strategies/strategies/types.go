package strategies

import (
	"TradingBot/src/markets/baseMarketClass"
	"TradingBot/src/types"
)

type StrategyParams struct {
	BaseMarketClass       *baseMarketClass.BaseMarketClass
	MarketStrategyParams  *types.MarketStrategyParams
	WithPendingOrders     bool
	CloseOrdersOnBadTrend bool
}
