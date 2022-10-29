package main

import (
	"TradingBot/src/manager"
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/simulator"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
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

	container := services.GetServicesContainer()
	container.Initialize(true)

	brokerAPI = getAPIInstance(container)
	if brokerAPI == nil {
		panic("Failed to get the broker API instance")
	}

	container.SetAPI(brokerAPI)

	_, err := brokerAPI.Login()
	if err != nil {
		fmt.Printf("%#v", "Login error -> "+err.Error())
		return
	}

	setupOSSignalsNotifications(brokerAPI)

	var waitingGroup sync.WaitGroup
	waitingGroup.Add(1)

	manager := &manager.Manager{
		ServicesContainer: container,
	}
	manager.Run()

	// Wait forever, this bot may never die
	waitingGroup.Wait()
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

func getAPIInstance(container *services.Container) api.Interface {
	user, password, accountID, apiName := getEnvVars()

	apis := make(map[string]func(
		credentials *api.Credentials,
		httpClient httpclient.Interface,
		logger logger.Interface,
	) api.Interface)

	credentials := &api.Credentials{
		Username:  user,
		Password:  password,
		AccountID: accountID,
	}

	apis["simulator"] = simulator.CreateAPIServiceInstance
	apis["ibroker"] = ibroker.CreateAPIServiceInstance

	return apis[apiName](
		credentials,
		container.HttpClient,
		container.Logger,
	)
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
