package logger

import "TradingBot/src/services/logger/types"

// Interface - For logging messages with different log levels
type Interface interface {
	Log(message string, logType ...types.LogType)
	Error(message string, logType ...types.LogType)
	ResetLogs()
}
