package strategies

import (
	"TradingBot/src/markets/interfaces"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
	"TradingBot/src/types"
)

type StrategyParams struct {
	MarketStrategyParams *types.MarketStrategyParams

	MarketData              *types.MarketData
	APIData                 api.DataInterface
	CandlesHandler          candlesHandler.Interface
	TrendsService           trends.Interface
	HorizontalLevelsService horizontalLevels.Interface

	API            api.Interface
	APIRetryFacade retryFacade.Interface

	Market interfaces.MarketInterface
}
