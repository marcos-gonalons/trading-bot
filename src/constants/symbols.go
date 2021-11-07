package constants

import (
	"TradingBot/src/services/api/ibroker"
	logger "TradingBot/src/services/logger/types"
	"TradingBot/src/types"
)

// TradingHours right now must be in Spanish time
// They must be changed every summer/winter accordingly

// Start hour is included, end hour is excluded
// For example, from 7 to 22, it will execute trades from 7 to 8, but not from 22 to 23.

var ForexType types.MarketType = "forex"
var IndexType types.MarketType = "index"

var Symbols = []types.Symbol{
	{
		BrokerAPIName: ibroker.GER30SymbolName,
		SocketName:    "FX:GER30",
		PriceDecimals: 1,
		TradingHours: types.TradingHours{
			Start: 7,
			End:   22,
		},
		TradeableOnWeekends: false,
		MaxSpread:           4,
		LogType:             logger.GER30,
		MarketType:          IndexType,
	},
	{
		BrokerAPIName: ibroker.EURUSDSymbolName,
		SocketName:    "FX:EURUSD",
		PriceDecimals: 5,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           999999,
		LogType:             logger.EURUSD,
		MarketType:          ForexType,
	},
	{
		BrokerAPIName: ibroker.GBPUSDSymbolName,
		SocketName:    "FX:GBPUSD",
		PriceDecimals: 5,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           999999,
		LogType:             logger.GBPUSD,
		MarketType:          ForexType,
	},
	{
		BrokerAPIName: ibroker.USDCADSymbolName,
		SocketName:    "FX:USDCAD",
		PriceDecimals: 5,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           999999,
		LogType:             logger.USDCAD,
		MarketType:          ForexType,
	},
	{
		BrokerAPIName: ibroker.USDJPYSymbolName,
		SocketName:    "FX:USDJPY",
		PriceDecimals: 3,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           999999,
		LogType:             logger.USDJPY,
		MarketType:          ForexType,
	},
	{
		BrokerAPIName: ibroker.USDCHFSymbolName,
		SocketName:    "FX:USDCHF",
		PriceDecimals: 5,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           999999,
		LogType:             logger.USDCHF,
		MarketType:          ForexType,
	},
	{
		BrokerAPIName: ibroker.NZDUSDSymbolName,
		SocketName:    "FX:NZDUSD",
		PriceDecimals: 5,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           999999,
		LogType:             logger.NZDUSD,
		MarketType:          ForexType,
	},
	{
		BrokerAPIName: ibroker.AUDUSDSymbolName,
		SocketName:    "FX:AUDUSD",
		PriceDecimals: 5,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           999999,
		LogType:             logger.AUDUSD,
		MarketType:          ForexType,
	},
}
