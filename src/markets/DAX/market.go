package DAX

import (
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
	tradingHoursUTC[0] = []int{21, 22, 23}

	market.MarketData = types.MarketData{
		BrokerAPIName: ibroker.GER30SymbolName,
		SocketName:    "FX:GER30",
		PriceDecimals: 1,
		TradingHours:  tradingHoursUTC,
		MaxSpread:     4,
		LogType:       loggerTypes.GER30,
		Rollover:      0,
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
		CandlesFileName:        "DAX-4H.csv",
		LongSetupParams:        &EMACrossoverLongParams,
		ShortSetupParams:       &EMACrossoverShortParams,
		EurExchangeRate:        1,
		PositionSizeMultiplier: .5,
		MinPositionSize:        1,
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
