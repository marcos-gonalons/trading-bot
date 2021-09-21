package logger

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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

// Logger ...
type Logger struct {
	rootPath  string
	fileNames map[LogType]string
	mtx       sync.Mutex
}

// Error logs the error in the normal log file + also in the error log file
func (logger *Logger) Error(message string, logType ...LogType) {
	logType = append(logType, ErrorLog)
	logger.Log(message, logType...)
}

// Log logs a message
func (logger *Logger) Log(message string, logType ...LogType) {
	var logFileName string
	isError := false

	if len(logType) > 0 {
		logFileName = logger.fileNames[logType[0]]
		if logType[len(logType)-1] == ErrorLog {
			isError = true
		}
	} else {
		logFileName = logger.fileNames[Default]
	}

	if logFileName == logger.fileNames[Default] || isError {
		var ioWriter io.Writer
		ioWriter = os.Stdout
		fmt.Fprintf(ioWriter, message)
		fmt.Fprintf(ioWriter, "\n\n")
	}

	logger.doLog(message, logFileName)
	if isError {
		logger.doLog(message, logger.fileNames[ErrorLog])
	}
}

// ResetLogs ...
func (logger *Logger) ResetLogs() {
	directory := logger.rootPath
	now := time.Now()

	osDir, err := os.Open(directory)
	if err != nil {
		panic("Error opening the directory" + directory + " -> " + err.Error())
	}
	files, err := osDir.Readdir(0)
	if err != nil {
		panic("Error reading the directory" + directory + " -> " + err.Error())
	}

	bkFolder := directory + "/backup-" + now.Format("2006-01-02")
	err = os.Mkdir(bkFolder, os.ModePerm)
	if err != nil {
		panic("Error while creating the backup log folder - " + bkFolder)
	}

	for index := range files {
		file := files[index]
		err = os.Rename(directory+"/"+file.Name(), bkFolder+"/"+file.Name())
		if err != nil {
			panic("Error moving the log file to the backup folder -> " + bkFolder + " -> " + file.Name() + " -> " + err.Error())
		}

		if strings.Contains(file.Name(), "backup") {
			if now.Unix()-file.ModTime().Unix() > 60*60*24*7 {
				err = os.RemoveAll("logs/" + file.Name())
				if err != nil {
					panic("Error deleting the old backup folder " + file.Name() + " -> " + err.Error())
				}
			}
		}
	}
}

func (logger *Logger) doLog(message string, logFileName string) {
	var now = time.Now()
	var folderPath = logger.rootPath
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		logger.mtx.Lock()
		err = os.Mkdir(folderPath, os.ModePerm)
		logger.mtx.Unlock()
		if err != nil {
			panic("Error while creating the log folder - " + folderPath)
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
			panic("Error while creating the log file - " + fullLogFilePath)
		}
	}

	logger.mtx.Lock()
	_, err = logFile.Write(getFormattedMessage(message, now))
	logger.mtx.Unlock()

	if err != nil {
		panic("Error while writting to log file - " + logFile.Name())
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
	filePathsMap[GetStateRequest] = "getStateRequest"
	filePathsMap[ModifyPositionRequest] = "modifyPositionRequest"
	filePathsMap[ErrorLog] = "errors"
	filePathsMap[GER30] = "GER30"
	filePathsMap[EURUSD] = "EURUSD"
	filePathsMap[GBPUSD] = "GBPUSD"
	filePathsMap[USDCAD] = "USDCAD"
	filePathsMap[USDJPY] = "USDJPY"
	filePathsMap[USDCHF] = "USDCHF"
	filePathsMap[NZDUSD] = "NZDUSD"
	filePathsMap[AUDUSD] = "AUDUSD"

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
