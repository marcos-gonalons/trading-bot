package types

type Symbol struct {
	BrokerAPIName     string
	SocketName        string
	MaxSpread         float64
	PriceDecimals     int64
	TradingHours      TradingHours
	ValidTradingTimes ValidTradingTimes
}

type ValidTradingTimes struct {
	Longs  TradingTimes
	Shorts TradingTimes
}

type TradingTimes struct {
	ValidMonths    []string
	ValidWeekdays  []string
	ValidHalfHours []string
}

type TradingHours struct {
	Start uint
	End   uint
}
