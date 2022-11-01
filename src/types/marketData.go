package types

import (
	logger "TradingBot/src/services/logger/types"
)

type MarketData struct {
	BrokerAPIName          string
	SocketName             string
	PriceDecimals          int64
	TradingHours           map[int][]int // Weekday => array of hours
	MaxSpread              float64
	LogType                logger.LogType
	MarketType             MarketType
	Rollover               float64
	Timeframe              Timeframe
	CandlesFileName        string
	EurExchangeRate        float64
	LongSetupParams        *MarketStrategyParams
	ShortSetupParams       *MarketStrategyParams
	PositionSizeMultiplier float64
}

type MarketType string
