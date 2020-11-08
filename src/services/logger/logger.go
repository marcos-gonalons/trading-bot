package logger

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

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
)

// Logger ...
type Logger struct {
	rootPath  string
	fileNames map[LogType]string
	mtx       sync.Mutex
}

// Log logs a message
func (logger *Logger) Log(message string, logType ...LogType) {
	var ioWriter io.Writer

	ioWriter = os.Stdout
	fmt.Fprintf(ioWriter, message)
	fmt.Fprintf(ioWriter, "\n\n")

	var logFileName string
	if len(logType) > 0 {
		logFileName = logger.fileNames[logType[0]]
	} else {
		logFileName = logger.fileNames[Default]
	}

	var now = time.Now()
	var folderPath = logger.rootPath + strconv.Itoa(now.Year()) + "_" + now.Month().String()
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		logger.mtx.Lock()
		err = os.Mkdir(folderPath, os.ModePerm)
		logger.mtx.Unlock()
		if err != nil {
			fmt.Printf("%v", err)
			return
		}
	}

	var fullLogFilePath = folderPath + "/" + logFileName

	var logFile, err = os.OpenFile(fullLogFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	defer logFile.Close()
	if err != nil {
		logger.mtx.Lock()
		logFile, err = os.Create(fullLogFilePath)
		logger.mtx.Unlock()
		defer logFile.Close()
		if err != nil {
			fmt.Printf("%v", err)
			return
		}
	}

	logger.mtx.Lock()
	_, err = logFile.Write(getFormattedMessage(message, now))
	logger.mtx.Unlock()

	if err != nil {
		fmt.Printf("%v", err)
	}
}

var logger *Logger

func init() {
	filePathsMap := make(map[LogType]string)
	filePathsMap[Default] = "bot"
	filePathsMap[LoginRequest] = "loginRequest"
	filePathsMap[GetQuoteRequest] = "getQuoteRequest"
	filePathsMap[CreateOrderRequest] = "createOrderRequest"
	filePathsMap[GetOrdersRequest] = "getOrdersRequest"
	filePathsMap[ModifyOrderRequest] = "modifyOrderRequest"
	filePathsMap[CloseOrderRequest] = "closeOrderRequest"
	filePathsMap[GetPositionsRequest] = "getPositionsRequest"
	filePathsMap[ClosePositionRequest] = "closePositionRequest"

	logger = &Logger{
		"logs/",
		filePathsMap,
		sync.Mutex{},
	}
}

// GetInstance returns the logger instance
func GetInstance() *Logger {
	return logger
}

func getFormattedMessage(message string, now time.Time) []byte {
	message = strconv.Itoa(now.Day()) + " - " +
		getWithLeadingZero(now.Hour()) + ":" +
		getWithLeadingZero(now.Minute()) + ":" +
		getWithLeadingZero(now.Second()) + " - " + message

	return []byte(message + "\n")
}

func getWithLeadingZero(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}
