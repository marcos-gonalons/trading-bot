package loginshandler

import (
	"time"
)

// Login ...
func Login(loginFunc func() error) {
	err := loginFunc()
	if err != nil {
		go loginRetries(loginFunc)
	}
}

func loginRetries(loginFunc func() error) {
	var err error
	for {
		now := time.Now()

		if now.Format("15") == "03" {
			panic("Too many failed login attempts. Last error was " + err.Error())
		}
		err = loginFunc()

		if err == nil {
			break
		}

		time.Sleep(1 * time.Minute)
	}
}
