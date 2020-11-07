package logger

// Interface - For logging messages with different log levels
type Interface interface {
	Log(message string, logType ...LogType)
}
