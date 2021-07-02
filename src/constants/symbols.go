package constants

import (
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/types"
)

// TradingHours right now must be in Spanish time
// They must be changed every summer/winter accordingly

// Start hour is included, end hour is excluded
// For example, from 7 to 22, it will execute trades from 7 to 8, but not from 22 to 23.

var Symbols = []types.Symbol{
	{
		BrokerAPIName: ibroker.GER30SymbolName,
		SocketName:    "FX:GER30",
		MaxSpread:     3,
		PriceDecimals: 1,
		TradingHours: types.TradingHours{
			Start: 7,
			End:   22,
		},
	},
	{
		BrokerAPIName: ibroker.SP500SymbolName,
		SocketName:    "TODO-TODO",
		MaxSpread:     3,
		PriceDecimals: 2,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
	},
	{
		BrokerAPIName: "__test__",
		SocketName:    "BINANCE:BTCUSD",
		MaxSpread:     200,
		PriceDecimals: 1,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
	},
}