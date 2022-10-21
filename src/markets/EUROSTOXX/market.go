package EUROSTOXX

import (
	"TradingBot/src/constants"
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
		BrokerAPIName: ibroker.EUROSTOXXSymbolName,
		SocketName:    "FXOPEN:ESX50",
		PriceDecimals: 1,
		TradingHours: types.TradingHours{
			Start: 8,
			End:   22,
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
		CandlesFileName:  "EUROSTOXX-4H.csv",
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
