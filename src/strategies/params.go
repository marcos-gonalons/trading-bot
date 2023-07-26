package strategies

import (
	"TradingBot/src/markets"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/types"
)

type Params struct {
	Type                 string // long or short
	MarketStrategyParams *types.MarketStrategyParams

	MarketData     *types.MarketData
	CandlesHandler candlesHandler.Interface
	Market         markets.MarketInterface
}
