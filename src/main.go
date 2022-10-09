package main

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/api/simulator"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
)

func main() {
	var brokerAPI api.Interface

	// TODO: Add functionality to pause/resume the bot with os signals
	defer func() {
		panicCatcher(recover(), brokerAPI)
	}()

	brokerAPI = getAPIInstance()
	if brokerAPI == nil {
		panic("Failed to get the broker API instance")
	}

	_, err := brokerAPI.Login()
	if err != nil {
		fmt.Printf("%#v", "Login error -> "+err.Error())
		return
	}

	setupOSSignalsNotifications(brokerAPI)

	var waitingGroup sync.WaitGroup
	waitingGroup.Add(1)
	apiRetryFacade := &retryFacade.APIFacade{
		API:    brokerAPI,
		Logger: logger.GetInstance(),
	}
	handler := &strategies.Handler{
		Logger:         logger.GetInstance(),
		API:            brokerAPI,
		APIRetryFacade: apiRetryFacade,
		APIData:        &api.Data{},
	}
	handler.Run()
	waitingGroup.Wait() // Wait forever, this script should never die
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
		//API.CloseAllPositions()
		//API.CloseAllOrders()
		os.Exit(0)
	}()
}

func getAPIInstance() api.Interface {
	user, password, accountID, apiName := getEnvVars()

	apis := make(map[string]api.Interface)

	credentials := &api.Credentials{
		Username:  user,
		Password:  password,
		AccountID: accountID,
	}

	apis["simulator"] = simulator.CreateAPIServiceInstance()
	apis["ibroker"] = ibroker.CreateAPIServiceInstance(credentials)

	return apis[apiName]
}

func getEnvVars() (user, password, accountID, apiName string) {
	user = os.Getenv("USERNAME")
	password = os.Getenv("PASSWORD")
	accountID = os.Getenv("ACCOUNT_ID")
	apiName = os.Getenv("API_NAME")

	if apiName == "" {
		panic("empty API_NAME env var")
	}

	return
}
