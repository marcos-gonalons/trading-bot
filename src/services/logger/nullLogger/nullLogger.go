package logger

import (
	"TradingBot/src/services/logger/types"
)

type NullLogger struct{}

var nullLogger *NullLogger

func (logger *NullLogger) Error(message string, logType ...types.LogType) {
}

func (logger *NullLogger) Log(message string, logType ...types.LogType) {
}

func (logger *NullLogger) ResetLogs() {
}

func GetInstance() *NullLogger {
	return nullLogger
}
