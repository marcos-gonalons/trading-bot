package EURUSD

import (
	"TradingBot/src/markets"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	loggerTypes "TradingBot/src/services/logger/types"
	"TradingBot/src/strategies"
	"TradingBot/src/types"
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
		TradingHours: types.TradingHours{
			Start: 0,
			End:   0,
		},
		TradeableOnWeekends: false,
		MaxSpread:           999999,
		LogType:             loggerTypes.EURUSD,
		MarketType:          "forex", // todo: move to constant. we have 'forex' and 'index' for now
		Timeframe: types.Timeframe{
			Value: 4,
			Unit:  "h",
		},
		CandlesFileName:  "EURUSD-4H.csv",
		Rollover:         .7, // Only used in market replay command
		LongSetupParams:  &EMACrossoverLongParams,
		ShortSetupParams: &EMACrossoverShortParams,
		EurExchangeRate:  1,
	}

	market.ToExecuteOnNewCandle = market.GetFuncToExecuteOnNewCandle()

	return market
}

func (s *Market) GetFuncToExecuteOnNewCandle() func() {
	return func() {
		s.Log("Calling resistanceBounce strategy")
		strategies.ResistanceBounce(strategies.StrategyParams{
			MarketStrategyParams:    &EMACrossoverLongParams,
			MarketData:              &s.MarketData,
			APIData:                 s.APIData,
			CandlesHandler:          s.CandlesHandler,
			TrendsService:           s.TrendsService,
			HorizontalLevelsService: s.HorizontalLevelsService,
			API:                     s.API,
			APIRetryFacade:          s.APIRetryFacade,
			Market:                  s,
		})

		/*
			s.Log("Calling supportBounce strategy")
			strategies.SupportBounce(strategies.StrategyParams{})
		*/
	}
}
