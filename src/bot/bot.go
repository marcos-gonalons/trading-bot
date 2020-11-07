package bot

import (
	"TradingBot/src/bot/loginshandler"
	"TradingBot/src/services/api"
	"fmt"
	"strconv"
	"time"
)

var previousExecutionTime time.Time
var failedGetQuoteRequestsInARow int = 0

// Execute - This is the code that is executed every 1.66666 seconds in an infinite loop
func Execute(API api.Interface) {
	now := time.Now()

	currentHour, _ := strconv.Atoi(now.Format("15"))
	previousHour, _ := strconv.Atoi(previousExecutionTime.Format("15"))

	if currentHour == 2 && previousHour == 1 {
		loginshandler.HandleLogins(API.Login)
	}

	if currentHour < 6 || currentHour > 21 {
		return
	}

	if failedGetQuoteRequestsInARow == 100 {
		// TODO: Close all limit/stop orders here
		panic("There is something wrong when fetching the quotes")
	}

	var err error

	quote, err := API.GetQuote("GER30")
	if err != nil {
		failedGetQuoteRequestsInARow++
		return
	}
	failedGetQuoteRequestsInARow = 0

	fmt.Printf("\n\n\n%#v\n\n\n", quote)

	previousExecutionTime = now
}
