package logger

import (
	"TradingBot/src/services/logger/types"
	"TradingBot/src/utils"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const DAYS_TO_KEEP_OLD_LOGS = 5

// Logger ...
type Logger struct {
	rootPath  string
	fileNames map[types.LogType]string
	mtx       sync.Mutex
}

// Error logs the error in the normal log file + also in the error log file
func (logger *Logger) Error(message string, logType ...types.LogType) {
	logType = append(logType, types.ErrorLog)
	logger.Log(message, logType...)
}

// Log logs a message
func (logger *Logger) Log(message string, logType ...types.LogType) {
	// todo: maybe add back the go routine
	func() {
		var logFileName string
		isError := false

		if len(logType) > 0 {
			logFileName = logger.fileNames[logType[0]]
			if logType[len(logType)-1] == types.ErrorLog {
				isError = true
			}
		} else {
			logFileName = logger.fileNames[types.Default]
		}

		if logFileName == logger.fileNames[types.Default] || isError {
			var ioWriter io.Writer
			ioWriter = os.Stdout
			fmt.Fprintf(ioWriter, message)
			fmt.Fprintf(ioWriter, "\n\n")
		}

		logger.doLog(message, logFileName)
		if isError {
			logger.doLog(message, logger.fileNames[types.ErrorLog])
		}
	}()
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

	bkFolder := directory + "backup-" + now.Format("2006-01-02") + "-" + utils.GetRandomString(4)
	err = os.Mkdir(bkFolder, os.ModePerm)
	if err != nil {
		panic("Error while creating the backup log folder - " + bkFolder + ". Error is " + err.Error())
	}

	for index := range files {
		file := files[index]

		if !strings.Contains(file.Name(), "backup") {
			err = os.Rename(directory+file.Name(), bkFolder+"/"+file.Name())
			if err != nil {
				panic("Error moving the log file to the backup folder -> " + bkFolder + " -> " + file.Name() + " -> " + err.Error())
			}
		} else {
			if now.Unix()-file.ModTime().Unix() > 60*60*24*DAYS_TO_KEEP_OLD_LOGS {
				err = os.RemoveAll(logger.rootPath + file.Name())
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

	var fullLogFilePath = folderPath + logFileName

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
	filePathsMap := make(map[types.LogType]string)
	filePathsMap[types.Default] = "bot"
	filePathsMap[types.LoginRequest] = "loginRequest"
	filePathsMap[types.GetQuoteRequest] = "getQuoteRequest"
	filePathsMap[types.CreateOrderRequest] = "createOrderRequest"
	filePathsMap[types.GetOrdersRequest] = "getOrdersRequest"
	filePathsMap[types.ModifyOrderRequest] = "modifyOrderRequest"
	filePathsMap[types.CloseOrderRequest] = "closeOrderRequest"
	filePathsMap[types.GetPositionsRequest] = "getPositionsRequest"
	filePathsMap[types.ClosePositionRequest] = "closePositionRequest"
	filePathsMap[types.GetStateRequest] = "getStateRequest"
	filePathsMap[types.ModifyPositionRequest] = "modifyPositionRequest"
	filePathsMap[types.ErrorLog] = "errors"
	filePathsMap[types.GER30] = "DAX"
	filePathsMap[types.SP500] = "SP500"
	filePathsMap[types.EUROSTOXX] = "EUROSTOXX"
	filePathsMap[types.EURUSD] = "EURUSD"
	filePathsMap[types.GBPUSD] = "GBPUSD"
	filePathsMap[types.USDCAD] = "USDCAD"
	filePathsMap[types.USDJPY] = "USDJPY"
	filePathsMap[types.USDCHF] = "USDCHF"
	filePathsMap[types.NZDUSD] = "NZDUSD"
	filePathsMap[types.AUDUSD] = "AUDUSD"

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
