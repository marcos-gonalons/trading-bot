package loginshandler

import (
	"TradingBot/src/services/api"
	"time"
)

// HandleLogins ...
func HandleLogins(loginFunc func() (*api.AccessToken, error)) {
	_, err := loginFunc()
	if err != nil {
		go loginRetries(loginFunc)
	}
}

func loginRetries(loginFunc func() (*api.AccessToken, error)) {
	var err error
	for {
		now := time.Now()

		if now.Format("15") == "03" {
			panic("Too many failed login attempts. Last error was " + err.Error())
		}
		_, err = loginFunc()

		if err == nil {
			break
		}

		time.Sleep(1 * time.Minute)
	}
}
