package EURUSD

import (
	"TradingBot/src/markets/baseMarketClass"
	"TradingBot/src/markets/interfaces"
	ibroker "TradingBot/src/services/api/ibroker/constants"
	loggerTypes "TradingBot/src/services/logger/types"
	"TradingBot/src/types"
)

// Market ...
type Market struct {
	baseMarketClass.BaseMarketClass
}

func (s *Market) OnNewCandle() {
	s.OnNewCandle()

	// todo: call ema crossover strategy, longs and shorts
	/*
		s.Log("Calling resistanceBounce strategy")
		strategies.ResistanceBounce(strategies.StrategyParams{
			BaseMarketClass:       &s.BaseMarketClass,
			MarketStrategyParams:  &ResistanceBounceParams,
			WithPendingOrders:     false,
			CloseOrdersOnBadTrend: false,
		})

		s.Log("Calling supportBounce strategy")
		strategies.SupportBounce(strategies.StrategyParams{
			BaseMarketClass:       &s.BaseMarketClass,
			MarketStrategyParams:  &SupportBounceParams,
			WithPendingOrders:     false,
			CloseOrdersOnBadTrend: false,
		})
	*/
}

func GetMarketInstance() interfaces.MarketInterface {
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

	return market
}
