package DAX

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
		BrokerAPIName: ibroker.GER30SymbolName,
		SocketName:    "FX:GER30",
		PriceDecimals: 1,
		TradingHours: types.TradingHours{
			Start: 3,
			End:   22,
		},
		TradeableOnWeekends: false,
		MaxSpread:           4,
		LogType:             loggerTypes.GER30,
		MarketType:          constants.IndexType,
		Rollover:            0,
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
		CandlesFileName:  "DAX-4H.csv",
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
