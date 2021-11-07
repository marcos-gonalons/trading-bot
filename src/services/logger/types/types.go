package types

// LogType determines which log file to use
type LogType uint8

const (
	// Default 1
	Default LogType = 1

	// LoginRequest 2
	LoginRequest LogType = 2

	// GetQuoteRequest 3
	GetQuoteRequest LogType = 3

	// CreateOrderRequest 4
	CreateOrderRequest LogType = 4

	// GetOrdersRequest 5
	GetOrdersRequest LogType = 5

	// ModifyOrderRequest 6
	ModifyOrderRequest LogType = 6

	// CloseOrderRequest 7
	CloseOrderRequest LogType = 7

	// GetPositionsRequest 8
	GetPositionsRequest LogType = 8

	// ClosePositionRequest 9
	ClosePositionRequest LogType = 9

	// GetStateRequest 10
	GetStateRequest LogType = 10

	// ModifyPositionRequest 11
	ModifyPositionRequest LogType = 11

	// ErrorLog 99
	ErrorLog LogType = 99

	// GER30 100
	GER30 LogType = 100

	// EURUSD 101
	EURUSD LogType = 101

	// GBPUSD 102
	GBPUSD LogType = 102

	// USDCAD 103
	USDCAD LogType = 103

	// USDJPY 104
	USDJPY LogType = 104

	// USDCHF 105
	USDCHF LogType = 105

	// NZDUSD 106
	NZDUSD LogType = 106

	// AUDUSD 107
	AUDUSD LogType = 107
)
