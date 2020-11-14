package main

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "TradingBot/src/services/logger"
)

func main() {
	var ibrokerAPI api.Interface

	a := fmt.Sprintf("%.2f", 12.0)
	fmt.Printf("\n\n\n%#v\n\n\n", a)
	panic("ok")
	defer func() {
		panicCatcher(recover(), ibrokerAPI)
	}()

	user, password, accountID, err := getArgs(os.Args[1:])
	if err != nil {
		fmt.Printf("%#v", err.Error())
		return
	}

	ibrokerAPI = ibroker.CreateAPIServiceInstance(
		&api.Credentials{
			Username:  user,
			Password:  password,
			AccountID: accountID,
		},
	)
	setupOSSignalsNotifications(ibrokerAPI)

	_, err = ibrokerAPI.Login()
	if err != nil {
		fmt.Printf("%#v", "Login error -> "+err.Error())
		return
	}

	var waitingGroup sync.WaitGroup
	waitingGroup.Add(1)
	for _, strategy := range strategies.GetStrategies(ibrokerAPI, logger.GetInstance()) {
		go strategy.Execute()
	}
	waitingGroup.Wait() // Wait forever, this script should never die
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

func panicCatcher(err interface{}, API api.Interface) {
	if err == nil {
		return
	}

	logger.GetInstance().Log("PANIC - " + fmt.Sprintf("%#v", err))
	API.CloseEverything()
}

func setupOSSignalsNotifications(API api.Interface) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Printf("%#v", sig)
		API.CloseEverything()
		os.Exit(0)
	}()
}
