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
		ValidTradingTimes: types.ValidTradingTimes{
			Longs: types.TradingTimes{
				ValidMonths:    []string{"January", "February", "March", "April", "May", "June", "July", "August", "September"},
				ValidWeekdays:  []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"},
				ValidHalfHours: []string{"9:00", "9:30", "10:00", "10:30", "11:00", "11:30", "12:00", "12:30", "13:00", "13:30", "14:00", "16:00", "16:30", "17:00", "17:30", "20:00", "20:30"},
			},
			Shorts: types.TradingTimes{
				ValidMonths:    []string{"January", "March", "April", "May", "June", "August", "September", "October", "December"},
				ValidWeekdays:  []string{"Monday", "Tuesday", "Thursday", "Friday"},
				ValidHalfHours: []string{"8:00", "8:30", "9:00", "10:00", "10:30", "11:00", "11:30", "12:00", "12:30", "13:00", "14:00", "14:30", "15:00", "15:30", "16:00", "16:30", "17:00", "18:00"},
			},
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
