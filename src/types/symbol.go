package types

type Symbol struct {
	BrokerAPIName       string
	SocketName          string
	PriceDecimals       int64
	TradingHours        TradingHours
	TradeableOnWeekends bool
	MaxSpread           float64
}

type TradingHours struct {
	Start uint
	End   uint
}
