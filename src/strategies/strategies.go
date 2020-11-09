package strategies

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/logger"
	mainstrategy "TradingBot/src/strategies/main"
)

// GetStrategies ...
func GetStrategies(API api.Interface, logger logger.Interface) (strategies []Interface) {
	strategies = append(strategies, &mainstrategy.Strategy{
		API:    API,
		Logger: logger,
	})
	return
}
