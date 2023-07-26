package NZDUSD

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
		BrokerAPIName: ibroker.NZDUSDSymbolName,
		SocketName:    "FX:NZDUSD",
		PriceDecimals: 5,
		TradingHours:  utils.GetForexUTCTradingHours(),
		MaxSpread:     999999,
		LogType:       loggerTypes.NZDUSD,
		Rollover:      .7,
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
		CandlesFileName:        "NZDUSD-4H.csv",
		EurExchangeRate:        1,
		PositionSizeMultiplier: 1,
		MinPositionSize:        10000,
		EmaCrossoverSetup: &types.SetupParams{
			LongSetupParams: &EMACrossoverLongParams,
		},
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
		if s.MarketData.EmaCrossoverSetup == nil {
			return
		}

		if s.MarketData.EmaCrossoverSetup.LongSetupParams != nil {
			s.Log("Calling EmaCrossoverLongs strategy")
			emaCrossover.EmaCrossoverLongs(strategies.Params{
				Type:                 ibroker.LongSide,
				MarketStrategyParams: s.MarketData.EmaCrossoverSetup.LongSetupParams,
				MarketData:           &s.MarketData,
				CandlesHandler:       s.CandlesHandler,
				Market:               s,
				Container:            s.Container,
			})
		}
	}
}
