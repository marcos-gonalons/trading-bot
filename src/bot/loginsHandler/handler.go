package loginshandler

import (
	"TradingBot/src/services/api"
	"time"
)

// Login - Refresh the API access token
func Login(API api.Interface, maxRetries uint, timeBetweenRetries time.Duration) {
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
