package types

type Symbol struct {
	BrokerAPIName string
	SocketName    string
	MaxSpread     float64
	PriceDecimals int64
	TradingHours  TradingHours
}

type TradingHours struct {
	Start uint
	End   uint
}
