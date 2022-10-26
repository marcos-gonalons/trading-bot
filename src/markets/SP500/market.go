package SP500

import (
	"TradingBot/src/constants"
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	loggerTypes "TradingBot/src/services/logger/types"
	"TradingBot/src/strategies"
	"TradingBot/src/strategies/emaCrossover"
	"TradingBot/src/types"
)

type Market struct {
	markets.BaseMarketClass
}

func GetMarketInstance() markets.MarketInterface {
	market := &Market{}

	market.MarketData = types.MarketData{
		BrokerAPIName: ibroker.SP500SymbolName,
		SocketName:    "VANTAGE:SP500",
		PriceDecimals: 1,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   23,
		},
		TradeableOnWeekends: false,
		MaxSpread:           4,
		LogType:             loggerTypes.SP500,
		MarketType:          constants.IndexType,
		Rollover:            0,
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
		CandlesFileName:        "SP500-4H.csv",
		LongSetupParams:        &EMACrossoverLongParams,
		ShortSetupParams:       &EMACrossoverShortParams,
		EurExchangeRate:        1,
		PositionSizeMultiplier: 2,
	}

	market.ToExecuteOnNewCandle = market.GetFuncToExecuteOnNewCandle()

	return market
}

func (s *Market) GetFuncToExecuteOnNewCandle() func() {
	return func() {
		s.Log("Calling EmaCrossoverLongs strategy")
		emaCrossover.EmaCrossoverLongs(strategies.Params{
			Type:                 ibroker.LongSide,
			MarketStrategyParams: &EMACrossoverLongParams,
			MarketData:           &s.MarketData,
			CandlesHandler:       s.CandlesHandler,
			Market:               s,
			Container:            s.Container,
		})

		// No shorts for SP500. Combining longs and shorts not very worth it for this market.
		// Better just stick with longs.
		/*
			s.Log("Calling EmaCrossoverShorts strategy")
			emaCrossover.EmaCrossoverShorts(strategies.Params{
				Type:                 ibroker.ShortSide,
				MarketStrategyParams: &EMACrossoverShortParams,
				MarketData:           &s.MarketData,
				CandlesHandler:       s.CandlesHandler,
				Market:               s,
				Container:            s.Container,
			})
		*/
	}
}
