package bot

import (
	"TradingBot/src/bot/loginshandler"
	"TradingBot/src/services/api"
	"time"
)

var previousExecutionTime time.Time

// Execute - This is the code that is executed every 1.66666 seconds in an infinite loop
func Execute(API api.Interface) {

	now := time.Now()

	loginshandler.HandleLogins(API.Login, now, previousExecutionTime)

	/**
		The bot will be active between 06:00 to 21:30
		if hour between 6 and 21:30, let's go
	**/

	previousExecutionTime = now
}
