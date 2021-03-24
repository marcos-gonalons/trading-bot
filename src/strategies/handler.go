package strategies

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler"
	"TradingBot/src/services/logger"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/strategies/breakoutAnticipation"
	"TradingBot/src/utils"
	"net"
	"strconv"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper"
)

// Handler ...
type Handler struct {
	API            api.Interface
	APIRetryFacade retryFacade.Interface
	Logger         logger.Interface

	strategies []Interface
	socket     tradingviewsocket.SocketInterface

	fetchError        error
	failedAPIRequests uint

	symbol string
}

// Run ...
func (s *Handler) Run() {
	s.symbol = ibroker.GER30SymbolName

	s.strategies = s.getStrategies()
	for _, strategy := range s.strategies {
		strategy.Initialize()
	}

	s.initSocket()

	go s.resetAtTwoAm()
	go s.checkSessionDisconnectedError()
	go s.fetchDataLoop()
	go s.panicIfTooManyAPIFails()
}

func (s *Handler) initSocket() {
	s.Logger.Log("Initializing the socket ...")
	tradingviewsocket, err := tradingviewsocket.Connect(
		s.onReceiveMarketData,
		s.onSocketError,
	)
	if err != nil {
		panic("Error while initializing the trading view socket -> " + err.Error())
	}

	var symbols = []string{"FX:GER30"}
	for _, symbol := range symbols {
		err = tradingviewsocket.AddSymbol(symbol)
		if err != nil {
			panic("Error while adding the symbol -> " + err.Error())
		}
	}

	s.socket = tradingviewsocket
}

func (s *Handler) onReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.Logger.Log("Received data -> " + symbol + " -> " + utils.GetStringRepresentation(data))

	for _, strategy := range s.strategies {
		// TODO: Only call the strategy if the data is received for the symbol that the strategy expects.
		// the breakout strategy that we have right now must only process FX:GER30 data
		// So do that somehow
		// Maybe each startegy will have "validSymbols" so I can check that from here
		go strategy.OnReceiveMarketData(symbol, data)
	}
}

func (s *Handler) onSocketError(err error) {
	s.Logger.Log("Socket error -> " + err.Error())
	s.socket.Close()
	s.initSocket()
}

func (s *Handler) resetAtTwoAm() {
	for {
		currentHour, _ := strconv.Atoi(time.Now().Format("15"))

		if currentHour == 2 {
			s.Logger.ResetLogs()

			err := s.socket.Close()
			if err != nil {
				s.Logger.Error("Error when restarting the socket -> " + err.Error())
			}

			for _, strategy := range s.strategies {
				go strategy.Reset()
			}

			s.Logger.Log("Refreshing access token by calling API.Login")
			s.APIRetryFacade.Login(retryFacade.RetryParams{
				DelayBetweenRetries: 30 * time.Second,
				MaxRetries:          120,
			})
		}

		time.Sleep(60 * time.Minute)
	}
}

func (s *Handler) fetchDataLoop() {
	for {
		var waitingGroup sync.WaitGroup
		now := time.Now()
		currentHour, _ := strconv.Atoi(now.Format("15"))
		if currentHour >= 8 && currentHour <= 21 {
			fetchFuncs := []func(){
				func() {
					quote := s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						return s.API.GetQuote(s.symbol)
					}).(*api.Quote)
					if quote != nil {
						for _, strategy := range s.strategies {
							strategy.SetCurrentBrokerQuote(quote)
						}
					}
				},
				func() {
					orders := s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						return s.API.GetOrders()
					}).([]*api.Order)
					if orders != nil {
						for _, strategy := range s.strategies {
							strategy.SetOrders(orders)
						}
					}
				},
				func() {
					positions := s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						return s.API.GetPositions()
					}).([]*api.Position)
					if positions != nil {
						for _, strategy := range s.strategies {
							strategy.SetPositions(positions)
						}
					}
				},
				func() {
					state := s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						return s.API.GetState()
					}).(*api.State)
					if state != nil {
						for _, strategy := range s.strategies {
							strategy.SetState(state)
						}
					}
				},
			}

			waitingGroup.Add(len(fetchFuncs))
			for _, fetchFunc := range fetchFuncs {
				go fetchFunc()
			}
		} else {
			waitingGroup.Add(1)
			waitingGroup.Done()
		}
		waitingGroup.Wait()
		time.Sleep(1666666 * time.Microsecond)
	}
}

func (s *Handler) fetch(fetchFunc func() (interface{}, error)) (result interface{}) {
	result, err := fetchFunc()

	if err == nil {
		s.fetchError = nil
		return
	}

	s.fetchError = err
	s.Logger.Error("Error while fetching data -> " + err.Error())
	s.failedAPIRequests++

	if err, isNetError := err.(net.Error); isNetError && err.Timeout() {
		currentTimeout := s.API.GetTimeout()
		if currentTimeout < 60*time.Second {
			s.Logger.Log("Increasing timeout ...")
			s.API.SetTimeout(currentTimeout + 10*time.Second)
		} else {
			s.Logger.Error("The API has a timeout problem")
		}
	} else {
		s.Logger.Log("Setting default timeout of 10 seconds ...")
		s.API.SetTimeout(10 * time.Second)
	}
	return
}

func (s *Handler) checkSessionDisconnectedError() {
	for {
		if s.API.IsSessionDisconnectedError(s.fetchError) {
			s.Logger.Log("Session is disconnected. Loggin in again ... ")
			s.APIRetryFacade.Login(retryFacade.RetryParams{
				DelayBetweenRetries: 0,
				MaxRetries:          0,
			})
		}
		s.fetchError = nil
		time.Sleep(5 * time.Second)
	}
}

func (s *Handler) panicIfTooManyAPIFails() {
	for {
		if s.failedAPIRequests >= 50 {
			panic("There is something wrong with the API - Check logs - Stopping bot")
		}
		s.failedAPIRequests = 0
		time.Sleep(1 * time.Minute)
	}
}

func (s *Handler) getStrategies() []Interface {
	breakoutAnticipationStrategy := breakoutAnticipation.GetStrategyInstance(
		s.APIRetryFacade,
		s.Logger,
		s.symbol,
	)
	candlesHandler := &candlesHandler.Service{
		Logger:    s.Logger,
		Symbol:    s.symbol,
		Timeframe: breakoutAnticipationStrategy.Timeframe,
	}
	breakoutAnticipationStrategy.SetCandlesHandler(candlesHandler)
	breakoutAnticipationStrategy.SetHorizontalLevelsService(horizontalLevels.GetServiceInstance(candlesHandler))

	return []Interface{breakoutAnticipationStrategy}
}
