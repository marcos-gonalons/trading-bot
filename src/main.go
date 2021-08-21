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
	"runtime/debug"
	"sync"
	"syscall"

	_ "TradingBot/src/services/logger"
)

func main() {
	var ibrokerAPI api.Interface

	// TODO: Add functionality to pause/resume the bot with os signals
	defer func() {
		panicCatcher(recover(), ibrokerAPI)
	}()

	user, password, accountID, apiURL, err := getEnvVars()
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
		apiURL,
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

func getEnvVars() (user, password, accountID, apiURL string, err error) {
	user = os.Getenv("USERNAME")
	password = os.Getenv("PASSWORD")
	accountID = os.Getenv("ACCOUNT_ID")
	apiURL = os.Getenv("API_URL")

	if user == "" {
		err = errors.New("empty USER env var")
	}
	if password == "" {
		err = errors.New("empty PASSWORD env var")
	}
	if accountID == "" {
		err = errors.New("empty ACCOUNT_ID env var")
	}
	if apiURL == "" {
		err = errors.New("empty API_URL env var")
	}

	return
}

func panicCatcher(err interface{}, API api.Interface) {
	if err == nil {
		return
	}

	// TODO: Mobile alert or something.

	logger.GetInstance().Log("PANIC - " + fmt.Sprintf("%#v", err))
	logger.GetInstance().Log("\n" + string(debug.Stack()))

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
