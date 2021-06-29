package types

type Symbol struct {
	BrokerAPIName string
	SocketName    string
	TradingHours  TradingHours
}

type TradingHours struct {
	Start uint
	End   uint
}
