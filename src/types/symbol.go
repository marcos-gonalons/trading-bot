package types

import logger "TradingBot/src/services/logger/types"

type Symbol struct {
	BrokerAPIName       string
	SocketName          string
	PriceDecimals       int64
	TradingHours        TradingHours
	TradeableOnWeekends bool
	MaxSpread           float64
	LogType             logger.LogType
	MarketType          MarketType
	Rollover            float64
}
type MarketType string

type TradingHours struct {
	Start uint
	End   uint
}
