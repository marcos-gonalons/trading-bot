package strategies

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies/markets/interfaces"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"net"
	"strings"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

const DailyResetHour = 2

// Handler ...
type Handler struct {
	API            api.Interface
	APIData        api.DataInterface
	APIRetryFacade retryFacade.Interface
	Logger         logger.Interface

	markets []interfaces.MarketInterface
	socket  tradingviewsocket.SocketInterface

	fetchError        error
	failedAPIRequests uint

	symbolsForAPI    []*types.Symbol
	symbolsForSocket []*types.Symbol
}

// Run ...
func (s *Handler) Run() {

	s.markets = s.GetMarkets()

	s.initSymbolsArrays()
	s.initMarkets()
	s.initSocket()

	go s.dailyReset()
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

	s.Logger.Log("Adding symbols to socket " + utils.GetStringRepresentation(s.symbolsForSocket))
	for _, symbol := range s.symbolsForSocket {
		s.Logger.Log("Adding symbol to socket ... " + symbol.SocketName)
		err = tradingviewsocket.AddSymbol(symbol.SocketName)
		if err != nil {
			panic("Error while adding the symbol -> " + err.Error())
		}
	}

	s.socket = tradingviewsocket
}

func (s *Handler) onReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.Logger.Log("Received data -> " + symbol + " -> " + utils.GetStringRepresentation(data))

	for _, market := range s.markets {
		if symbol != market.Parent().GetSymbol().SocketName {
			continue
		}
		if market.Parent().GetCurrentBrokerQuote() != nil {
			go market.OnReceiveMarketData(symbol, data)
		}
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

func (s *Handler) dailyReset() {
	for {
		currentHour, _ := utils.GetCurrentTimeHourAndMinutes()

		if currentHour == DailyResetHour {
			s.Logger.ResetLogs()

			err := s.socket.Close()
			if err != nil {
				s.Logger.Error("Error when restarting the socket -> " + err.Error())
			}

			for _, market := range s.markets {
				go market.DailyReset()
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

		for _, symbol := range s.symbolsForAPI {
			if !utils.IsNowWithinTradingHours(symbol) {
				continue
			}

			fetchFuncs = append(fetchFuncs,
				func(symbol string) func() {
					return func() {
						quote := s.fetch(func() (interface{}, error) {
							defer waitingGroup.Done()
							s.Logger.Log("Fetching quote ...")
							return s.API.GetQuote(symbol)
						}).(*api.Quote)
						s.Logger.Log("Quote is -> " + utils.GetStringRepresentation(quote))
						if quote != nil {
							for _, market := range s.markets {
								if market.Parent().GetSymbol().BrokerAPIName != symbol {
									continue
								}
								market.Parent().SetCurrentBrokerQuote(quote)
							}
						}
					}
				}(symbol.BrokerAPIName),
			)
		}

		if len(fetchFuncs) > 0 {
			fetchFuncs = append(fetchFuncs,
				func() {
					orders := s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						s.Logger.Log("Fetching orders ...")
						return s.API.GetOrders()
					}).([]*api.Order)
					s.Logger.Log("Orders is -> " + utils.GetStringRepresentation(orders))
					var o []*api.Order
					if orders != nil {
						o = orders
					}
					s.APIData.SetOrders(o)
				},
				func() {
					positions := s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						s.Logger.Log("Fetching positions ...")
						return s.API.GetPositions()
					}).([]*api.Position)
					s.Logger.Log("Positions is -> " + utils.GetStringRepresentation(positions))
					var p []*api.Position
					if positions != nil {
						p = positions
					}
					s.APIData.SetPositions(p)
				},
				func() {
					state := s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						s.Logger.Log("Fetching state ...")
						return s.API.GetState()
					}).(*api.State)
					s.Logger.Log("State is -> " + utils.GetStringRepresentation(state))
					if state != nil {
						s.APIData.SetState(state)
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
		// todo: adjust this
		if s.failedAPIRequests >= 500 {
			panic("There is something wrong with the API - Check logs - Stopping bot")
		}
		s.failedAPIRequests = 0
		time.Sleep(1 * time.Minute)
	}
}

func (s *Handler) initMarkets() {
	for _, market := range s.markets {
		s.Logger.Log("Initializing market " + market.Parent().GetSymbol().BrokerAPIName)
		go market.Initialize()
	}
}

func (s *Handler) initSymbolsArrays() {
	s.Logger.Log("Initializing symbols arrays ...")
	for _, market := range s.markets {
		symbol := market.Parent().GetSymbol()
		s.Logger.Log("Symbol: " + utils.GetStringRepresentation(symbol))

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
			if s.SocketName == symbol.SocketName {
				exists = true
				break
			}
		}
		if !exists {
			s.symbolsForSocket = append(s.symbolsForSocket, symbol)
		}
	}
}
