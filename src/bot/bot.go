package bot

import (
	"TradingBot/src/services/api"
)

// Execute - This is the code that is executed every second in an infinite loop
func Execute(API api.Interface) {
	API.Login()
}
