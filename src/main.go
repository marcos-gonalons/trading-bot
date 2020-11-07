package main

import (
	"TradingBot/src/bot"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"errors"
	"fmt"
	"os"
	"time"

	_ "TradingBot/src/services/logger"
)

func main() {
	user, password, accountID, err := getArgs(os.Args[1:])
	if err != nil {
		fmt.Printf("%#v", err.Error())
		return
	}

	ibrokerAPI := ibroker.CreateAPIServiceInstance(
		&api.Credentials{
			Username:  user,
			Password:  password,
			AccountID: accountID,
		},
	)
	_, err = ibrokerAPI.Login()
	if err != nil {
		fmt.Printf("%#v", "Login error -> "+err.Error())
		return
	}

	for {
		bot.Execute(ibrokerAPI)

		// Why 1.66666 seconds?
		// Tradingview sends the get quotes request once every 1.66666 seconds, so we should do the same.
		time.Sleep(1666666 * time.Microsecond)
	}
}

func getArgs(args []string) (user, password, accountID string, err error) {
	if len(args) == 3 {
		user = args[0]
		password = args[1]
		accountID = args[2]
	} else {
		err = errors.New("Need 3 arguments: username, password and accountID")
	}

	return
}
