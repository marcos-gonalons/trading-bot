package main

import (
	"TradingBot/src/bot"
	"TradingBot/src/services/api/ibroker"
	"errors"
	"fmt"
	"os"
	"time"
)

func main() {
	user, password, _, err := getArgs(os.Args[1:])

	if err != nil {
		fmt.Printf("%#v", err.Error())
		return
	}

	api := ibroker.CreateAPIServiceInstance()
	_, err = api.Login(user, password)
	if err != nil {
		fmt.Printf("Error while logging in\n" + err.Error())
		return
	}

	for {
		bot.Execute(api)

		// Why 1.66666 seconds?
		// Tradingview sends the get quotes request once every 1.66666 seconds, so we should do the same.
		time.Sleep(1.66666 * time.Second)
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
