package SP500

import (
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	loggerTypes "TradingBot/src/services/logger/types"
	"TradingBot/src/strategies/emaCrossover"
	"TradingBot/src/types"
)

type Market struct {
	markets.BaseMarketClass
}

func GetMarketInstance() markets.MarketInterface {
	market := &Market{}

	tradingHoursUTC := make(map[int][]int)

	// Monday
	tradingHoursUTC[1] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}
	// Tuesday
	tradingHoursUTC[2] = tradingHoursUTC[1]
	// Wednesday
	tradingHoursUTC[3] = tradingHoursUTC[1]
	// Thursday
	tradingHoursUTC[4] = tradingHoursUTC[1]
	// Friday
	tradingHoursUTC[5] = tradingHoursUTC[1]
	// Saturday
	tradingHoursUTC[6] = []int{}
	// Sunday
	tradingHoursUTC[0] = []int{23} // todo: summer time starts at 22 UTC, winter time starts at 23:00 UTC

	market.MarketData = types.MarketData{
		BrokerAPIName: ibroker.SP500SymbolName,
		SocketName:    "VANTAGE:SP500",
		PriceDecimals: 1,
		TradingHours:  tradingHoursUTC,
		MaxSpread:     4,
		LogType:       loggerTypes.SP500,
		Rollover:      0,
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
		CandlesFileName:        "SP500-4H.csv",
		EurExchangeRate:        1,
		PositionSizeMultiplier: 2,
		MinPositionSize:        1,
		EmaCrossoverSetup: &types.SetupParams{
			LongSetupParams:  &EMACrossoverLongParams,
			ShortSetupParams: &EMACrossoverShortParams,
		},
		SimulatorData: &types.SimulatorData{
			Spread:   2,
			Slippage: 2,
		},
	}

	market.ToExecuteOnNewCandle = market.GetFuncToExecuteOnNewCandle()

	return market
}

func (s *Market) GetFuncToExecuteOnNewCandle() func() {
	return func() {
		emaCrossover.OnNewCandle(s)
	}
}
