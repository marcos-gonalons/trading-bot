package EURUSD

import (
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	loggerTypes "TradingBot/src/services/logger/types"
	"TradingBot/src/strategies/ranges"
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
		EurExchangeRate:        1,
		PositionSizeMultiplier: 1,
		MinPositionSize:        10000,
		EmaCrossoverSetup: &types.SetupParams{
			LongSetupParams:  &EMACrossoverLongParams,
			ShortSetupParams: &EMACrossoverShortParams,
		},
		RangesSetup: &types.SetupParams{
			LongSetupParams:  &RangesLongParams,
			ShortSetupParams: &RangesShortParams,
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
		ranges.OnNewCandle(s)
		// emaCrossover.OnNewCandle(s)
	}
}
