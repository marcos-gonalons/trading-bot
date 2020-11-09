package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
	"strconv"
	"time"
)

var previousExecutionTime time.Time
var failedGetQuoteRequestsInARow int = 0

// Strategy ...
type Strategy struct {
	API    api.Interface
	Logger logger.Interface
}

// Execute ...
func (s *Strategy) Execute() {
	for {
		s.execute()
		// Why 1.66666 seconds?
		// Tradingview sends the get quotes request once every 1.66666 seconds, so we should do the same.
		time.Sleep(1666666 * time.Microsecond)
	}
}

func (s *Strategy) execute() {
	now := time.Now()

	currentHour, previousHour := s.getCurrentAndPreviousHour(now, previousExecutionTime)
	if currentHour == 2 && previousHour == 1 {
		s.Logger.ResetLogs()

		s.Logger.Log("Refreshing access token by calling API.Login")
		s.login(60, 1*time.Minute)
	}

	if currentHour < 6 || currentHour > 21 {
		s.Logger.Log("Doing nothing - Now it's not the time.")
		return
	}

	quote := s.getQuote("GER30")
	if quote == nil {
		return
	}

	/***
		When creating an order, I need to save the 3 created orders somewhere (the limit/stop order, it's sl and it's tp)
		The SL and the TP will have the parentID of the main one. The main one will have the parentID null
		All 3 orders will have the status "working".

		When modifying an order that hasn't been filled yet, I can use the ID of the main order to change it's sl, tp, or it's limit/stop price.
		When modifying the sl/tp of a position, I need to use the ID of the sl/tp order.




		Trading view does this API calls every time
		get quote
		get orders
		get state
		get positions

		So I should do the same

		Use goroutines + waiting group
		When all the requests are done; execute script code

		log something when it's time to execute the script code, so I can see how much time it takes to finish all 4 requests
	***/

	previousExecutionTime = now
}

func (s *Strategy) getCurrentAndPreviousHour(
	now time.Time,
	previous time.Time,
) (int, int) {
	currentHour, _ := strconv.Atoi(now.Format("15"))
	previousHour, _ := strconv.Atoi(previous.Format("15"))
	return currentHour, previousHour
}

func (s *Strategy) login(maxRetries uint, timeBetweenRetries time.Duration) {
	_, err := s.API.Login()
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
			_, err = s.API.Login()

			if err == nil {
				break
			}

			retriesAmount++
			time.Sleep(timeBetweenRetries)
		}
	}()
}

func (s *Strategy) getQuote(symbol string) *api.Quote {
	if failedGetQuoteRequestsInARow == 100 {
		panic("There is something wrong when fetching the quotes")
	}

	quote, err := s.API.GetQuote(symbol)
	if err != nil {
		failedGetQuoteRequestsInARow++
		s.Logger.Log("Error when fetching the quote - Fails in a row -> " + utils.IntToString(int64(failedGetQuoteRequestsInARow)))
		return nil
	}

	failedGetQuoteRequestsInARow = 0
	return quote
}
