package main

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies"
	"TradingBot/src/tradingviewsocket"
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

	/**
		fx va 1 pip por detras
		si fx:ger30 dice 13149.4, en ibroker es 13150.5

		leer con unauthorized de fx:ger30
		y tener en cuenta que va 1 por detras
	**/
	var waitingGroupX sync.WaitGroup
	waitingGroupX.Add(1)
	socket := tradingviewsocket.TradingviewSocket{
		OnReceiveMarketDataCallback: func(symbol string, data map[string]interface{}) {
			fmt.Printf("\n%#v\n", "received data")
			fmt.Printf("\n%#v\n", data)
		},
		OnErrorCallback: func(err error) {
			fmt.Printf("\n%#v\n", "error"+err.Error())
		},
	}

	/**
		El volumen se resetea cada dia a las 23:00 hora de espanya (al menos en eurusd)
		Y cuando se recibe el volumen se recibe el volumen acumulado desde el reseteo hasta ese momento.
	**/
	socket.AddSymbol("FX:GER30")
	socket.Init()
	waitingGroupX.Wait() // Wait forever, this script should never die

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
