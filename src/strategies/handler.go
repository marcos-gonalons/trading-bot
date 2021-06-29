package strategies

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
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

	symbolsForAPI    []*types.Symbol
	symbolsForSocket []*types.Symbol
}

// Run ...
func (s *Handler) Run() {

	s.initSymbolsArrays()
	s.initStrategies()
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

	for _, symbol := range s.symbolsForSocket {
		err = tradingviewsocket.AddSymbol(symbol.SocketName)
		if err != nil {
			panic("Error while adding the symbol -> " + err.Error())
		}
	}

	s.socket = tradingviewsocket
}

func (s *Handler) onReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.Logger.Log("Received data -> " + symbol + " -> " + utils.GetStringRepresentation(data))

	for _, strategy := range s.strategies {
		if symbol != strategy.GetSymbol().SocketName {
			continue
		}
		go strategy.OnReceiveMarketData(symbol, data)
	}
}

func (s *Handler) onSocketError(err error, context string) {
	s.Logger.Log("Socket error -> " + err.Error())
	s.Logger.Log("Context -> " + context)
	err = s.socket.Close()
	if err != nil {
		s.Logger.Error("Error when closing the socket -> " + err.Error())
		if !strings.Contains(err.Error(), "use of closed network connection") {
			return
		}
	}

	s.Logger.Log("Initializing the socket again ... ")
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
		var fetchFuncs []func()

		now := time.Now()
		currentHour, _ := strconv.Atoi(now.Format("15"))

		for _, symbol := range s.symbolsForAPI {
			if currentHour < int(symbol.TradingHours.Start) || currentHour >= int(symbol.TradingHours.End) {
				continue
			}

			fetchFuncs = append(fetchFuncs,
				func() {
					go func(symbol string) {
						quote := s.fetch(func() (interface{}, error) {
							defer waitingGroup.Done()
							return s.API.GetQuote(symbol)
						}).(*api.Quote)
						if quote != nil {
							for _, strategy := range s.strategies {
								if strategy.GetSymbol().BrokerAPIName != symbol {
									continue
								}
								strategy.SetCurrentBrokerQuote(quote)
							}
						}
					}(symbol.BrokerAPIName)
				},
			)
		}

		if len(fetchFuncs) > 0 {
			fetchFuncs = append(fetchFuncs,
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
					var p []*api.Position
					if positions != nil {
						p = positions
					}
					for _, strategy := range s.strategies {
						strategy.SetPositions(p)
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
			)
		}

		if len(fetchFuncs) == 0 {
			waitingGroup.Add(1)
			waitingGroup.Done()
		} else {
			waitingGroup.Add(len(fetchFuncs))
		}

		for _, fetchFunc := range fetchFuncs {
			go fetchFunc()
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

func (s *Handler) initStrategies() {
	s.strategies = s.getStrategies()
	for _, strategy := range s.strategies {
		go strategy.Initialize()
	}
}

func (s *Handler) initSymbolsArrays() {
	for _, strategy := range s.strategies {
		symbol := strategy.GetSymbol()

		exists := false
		for _, s := range s.symbolsForAPI {
			if s.BrokerAPIName == symbol.BrokerAPIName {
				exists = true
				break
			}
		}
		if !exists {
			s.symbolsForAPI = append(s.symbolsForAPI, symbol)
		}

		exists = false
		for _, s := range s.symbolsForSocket {
			if s.BrokerAPIName == symbol.SocketName {
				exists = true
				break
			}
		}
		if !exists {
			s.symbolsForSocket = append(s.symbolsForSocket, symbol)
		}
	}
}
