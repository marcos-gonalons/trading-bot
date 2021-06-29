package constants

import (
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/types"
)

// TradingHours right now must be in Spanish time
// They must be changed every summer/winter accordingly

var Symbols = []types.Symbol{
	{
		BrokerAPIName: ibroker.GER30SymbolName,
		SocketName:    "FX:GER30",
		TradingHours: types.TradingHours{
			Start: 7,
			End:   22,
		},
	},
	{
		BrokerAPIName: "__test__",
		SocketName:    "BINANCE:BTCUSD",
		TradingHours: types.TradingHours{
			Start: 0,
			End:   24,
		},
	},
}
