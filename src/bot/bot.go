package bot

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
	"os"
	"strconv"
	"time"
)

var previousExecutionTime time.Time
var failedGetQuoteRequestsInARow int = 0

// Execute - This is the code that is executed every 1.66666 seconds in an infinite loop
func Execute(API api.Interface, logger logger.Interface) {
	now := time.Now()

	currentHour, previousHour := getCurrentAndPreviousHour(now, previousExecutionTime)
	if currentHour == 2 && previousHour == 1 {
		resetLogs()

		logger.Log("Refreshing access token by calling API.Login")
		login(API, 60, 1*time.Minute)
	}

	if currentHour < 6 || currentHour > 21 {
		logger.Log("Doing nothing - Now it's not the time.")
		return
	}

	quote := getQuote("GER30", API, logger)
	if quote == nil {
		return
	}

	/***
		When creating an order, I need to save the 3 created orders somewhere (the limit/stop order, it's sl and it's tp)
		The SL and the TP will have the parentID of the main one. The main one will have the parentID null
		All 3 orders will have the status "working".

		When modifying an order that hasn't been filled yet, I can use the ID of the main order to change it's sl, tp, or it's limit/stop price.
		When modifying the sl/tp of a position, I need to use the ID of the sl/tp order.
	***/

	previousExecutionTime = now
}

func getCurrentAndPreviousHour(
	now time.Time,
	previous time.Time,
) (int, int) {
	currentHour, _ := strconv.Atoi(now.Format("15"))
	previousHour, _ := strconv.Atoi(previous.Format("15"))
	return currentHour, previousHour
}

func login(API api.Interface, maxRetries uint, timeBetweenRetries time.Duration) {
	_, err := API.Login()
	if err == nil {
		return
	}

	go func() {
		var err error
		var retriesAmount uint = 0
		for {
			if retriesAmount == maxRetries {
				panic("Too many failed login attempts. Last error was " + err.Error())
			}
			_, err = API.Login()

			if err == nil {
				break
			}

			retriesAmount++
			time.Sleep(timeBetweenRetries)
		}
	}()
}

func getQuote(
	symbol string,
	API api.Interface,
	logger logger.Interface,
) *api.Quote {
	if failedGetQuoteRequestsInARow == 100 {
		panic("There is something wrong when fetching the quotes")
	}

	quote, err := API.GetQuote(symbol)
	if err != nil {
		logger.Log("Error when fetchin the quote - Fails in a row -> " + utils.IntToString(int64(failedGetQuoteRequestsInARow)))
		failedGetQuoteRequestsInARow++
		return nil
	}

	failedGetQuoteRequestsInARow = 0
	return quote
}

func resetLogs() {
	directory := "logs"

	osDir, _ := os.Open(directory)
	files, _ := osDir.Readdir(0)

	for index := range files {
		file := files[index]
		os.Remove(directory + "/" + file.Name())
	}
}
