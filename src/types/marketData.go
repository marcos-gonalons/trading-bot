package types

import (
	logger "TradingBot/src/services/logger/types"
)

type SimulatorData struct {
	Spread   float64
	Slippage float64
}
type SetupParams struct {
	LongSetupParams  *MarketStrategyParams
	ShortSetupParams *MarketStrategyParams
}
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
	PositionSizeMultiplier float64
	MinPositionSize        int64
	SimulatorData          *SimulatorData

	EmaCrossoverSetup *SetupParams
	RangesSetup       *SetupParams
}

type MarketType string
