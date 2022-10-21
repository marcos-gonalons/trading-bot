package strategies

import (
	"TradingBot/src/markets"
	"TradingBot/src/services"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/types"
)

type StrategyParams struct {
	MarketStrategyParams *types.MarketStrategyParams

	MarketData     *types.MarketData
	CandlesHandler candlesHandler.Interface
	Market         markets.MarketInterface

	Container *services.Container
}
