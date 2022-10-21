package USDCAD

import (
	"TradingBot/src/constants"
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	loggerTypes "TradingBot/src/services/logger/types"
	"TradingBot/src/types"
)

type Market struct {
	markets.BaseMarketClass
}

func GetMarketInstance() markets.MarketInterface {
	market := &Market{}

	market.MarketData = types.MarketData{
		BrokerAPIName: ibroker.USDCADSymbolName,
		SocketName:    "FX:USDCAD",
		PriceDecimals: 5,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           999999,
		LogType:             loggerTypes.USDCAD,
		MarketType:          constants.ForexType,
		Rollover:            .7,
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
		CandlesFileName:  "USDCAD-4H.csv",
		LongSetupParams:  &EMACrossoverLongParams,
		ShortSetupParams: &EMACrossoverShortParams,
		EurExchangeRate:  1,
	}

	market.ToExecuteOnNewCandle = market.GetFuncToExecuteOnNewCandle()

	return market
}

func (s *Market) GetFuncToExecuteOnNewCandle() func() {
	return func() {
		/*
			s.Log("Calling supportBounce strategy")
			strategies.SupportBounce(strategies.StrategyParams{})
		*/
	}
}
