package EURUSD

import (
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	loggerTypes "TradingBot/src/services/logger/types"
	"TradingBot/src/strategies"
	"TradingBot/src/strategies/emaCrossover"
	"TradingBot/src/types"
	"TradingBot/src/utils"
)

type Market struct {
	markets.BaseMarketClass
}

func GetMarketInstance() markets.MarketInterface {
	market := &Market{}

	market.MarketData = types.MarketData{
		BrokerAPIName: ibroker.EURUSDSymbolName,
		SocketName:    "FX:EURUSD",
		PriceDecimals: 5,
		TradingHours:  utils.GetForexUTCTradingHours(),
		MaxSpread:     999999,
		LogType:       loggerTypes.EURUSD,
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
		CandlesFileName:        "EURUSD-4H.csv",
		Rollover:               .7,
		LongSetupParams:        &EMACrossoverLongParams,
		ShortSetupParams:       &EMACrossoverShortParams,
		EurExchangeRate:        1,
		PositionSizeMultiplier: 1,
		MinPositionSize:        10000,
		SimulatorData: &types.SimulatorData{
			Spread:   .00012,
			Slippage: .00012,
		},
	}

	market.ToExecuteOnNewCandle = market.GetFuncToExecuteOnNewCandle()

	return market
}

func (s *Market) GetFuncToExecuteOnNewCandle() func() {
	return func() {
		if s.MarketData.LongSetupParams != nil {
			s.Log("Calling EmaCrossoverLongs strategy")
			emaCrossover.EmaCrossoverLongs(strategies.Params{
				Type:                 ibroker.LongSide,
				MarketStrategyParams: s.MarketData.LongSetupParams,
				MarketData:           &s.MarketData,
				CandlesHandler:       s.CandlesHandler,
				Market:               s,
				Container:            s.Container,
			})
		}

		if s.MarketData.ShortSetupParams != nil {
			s.Log("Calling EmaCrossoverShorts strategy")
			emaCrossover.EmaCrossoverShorts(strategies.Params{
				Type:                 ibroker.ShortSide,
				MarketStrategyParams: s.MarketData.ShortSetupParams,
				MarketData:           &s.MarketData,
				CandlesHandler:       s.CandlesHandler,
				Market:               s,
				Container:            s.Container,
			})
		}
	}
}
