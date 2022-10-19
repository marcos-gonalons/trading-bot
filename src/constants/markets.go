package constants

import (
	ibroker "TradingBot/src/services/api/ibroker/constants"
	logger "TradingBot/src/services/logger/types"
	"TradingBot/src/types"
)

// TradingHours right now must be in Spanish time
// They must be changed every summer/winter accordingly

// Start hour is included, end hour is excluded
// For example, from 7 to 22, it will execute trades from 7 to 8, but not from 22 to 23.

const ForexType types.MarketType = "forex"
const IndexType types.MarketType = "index"

// todo: finish moving this data to markets folder

var Markets = []types.MarketData{
	{
		BrokerAPIName: ibroker.GER30SymbolName,
		SocketName:    "FX:GER30",
		PriceDecimals: 1,
		TradingHours: types.TradingHours{
			Start: 3,
			End:   22,
		},
		TradeableOnWeekends: false,
		MaxSpread:           4,
		LogType:             logger.GER30,
		MarketType:          IndexType,
		Rollover:            0,
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
	},

	/*
		{
			BrokerAPIName: ibroker.SP500SymbolName,
			SocketName:    "VANTAGE:SP500",
			PriceDecimals: 1,
			TradingHours: types.TradingHours{
				Start: 0,
				End:   23,
			},
			TradeableOnWeekends: false,
			MaxSpread:           4,
			LogType:             logger.SP500,
			MarketType:          IndexType,
			Rollover:            0,
			Timeframe: types.Timeframe{
				Value: 4,
				Unit:  "h",
			},
		},
		{
			BrokerAPIName: ibroker.EUROSTOXXSymbolName,
			SocketName:    "FXOPEN:ESX50",
			PriceDecimals: 1,
			TradingHours: types.TradingHours{
				Start: 8,
				End:   22,
			},
			TradeableOnWeekends: false,
			MaxSpread:           4,
			LogType:             logger.EUROSTOXX,
			MarketType:          IndexType,
			Rollover:            0,
			Timeframe: types.Timeframe{
				Value: 4,
				Unit:  "h",
			},
		},
		{

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
			Rollover:            .7,
			Timeframe: types.Timeframe{
				Value: 4,
				Unit:  "h",
			},
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
			Rollover:            .7,
			Timeframe: types.Timeframe{
				Value: 4,
				Unit:  "h",
			},
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
			Rollover:            0.6,
			Timeframe: types.Timeframe{
				Value: 4,
				Unit:  "h",
			},
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
			Rollover:            .7,
			Timeframe: types.Timeframe{
				Value: 4,
				Unit:  "h",
			},
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
			Rollover:            .7,
			Timeframe: types.Timeframe{
				Value: 4,
				Unit:  "h",
			},
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
			Rollover:            .7,
			Timeframe: types.Timeframe{
				Value: 4,
				Unit:  "h",
			},
		},
	*/
}
