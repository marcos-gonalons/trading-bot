package EUROSTOXX

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
		BrokerAPIName: ibroker.EUROSTOXXSymbolName,
		SocketName:    "FXOPEN:ESX50",
		PriceDecimals: 1,
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           4,
		LogType:             loggerTypes.EUROSTOXX,
		MarketType:          constants.IndexType,
		Rollover:            0,
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
		CandlesFileName:        "EUROSTOXX-4H.csv",
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
		s.Log("Calling EmaCrossoverShorts strategy")
		emaCrossover.EmaCrossoverShorts(strategies.Params{
			Type:                 ibroker.ShortSide,
			MarketStrategyParams: &EMACrossoverShortParams,
			MarketData:           &s.MarketData,
			CandlesHandler:       s.CandlesHandler,
			Market:               s,
			Container:            s.Container,
		})
	}
}
