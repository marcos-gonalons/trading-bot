package main

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
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

	// TODO: Add functionality to pause/resume the bot
	// Maybe with OS signals?
	// When receiving the specific signal, do not process the strategies code?
	// Can be a bit dangerous if a position is open or there is a pending order
	// But think further!

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
	apiRetryFacade := &retryFacade.APIFacade{
		API:    ibrokerAPI,
		Logger: logger.GetInstance(),
	}
	handler := &strategies.Handler{
		Logger:         logger.GetInstance(),
		API:            ibrokerAPI,
		APIRetryFacade: apiRetryFacade,
	}
	handler.Run()
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

	// TODO: Mobile alert or something.

	logger.GetInstance().Log("PANIC - " + fmt.Sprintf("%#v", err))

	API.CloseAllPositions()
	API.CloseAllOrders()
}

func setupOSSignalsNotifications(API api.Interface) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Printf("%#v", sig)
		API.CloseAllPositions()
		API.CloseAllOrders()
		os.Exit(0)
	}()
}
