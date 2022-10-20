package manager

import (
	"TradingBot/src/markets"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
	"net"
	"strings"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

const DailyResetHour = 2

// todo: all the services will be initialized here, example trends and horizontal levels service
type Manager struct {
	API            api.Interface
	APIData        api.DataInterface
	APIRetryFacade retryFacade.Interface
	Logger         logger.Interface

	markets []markets.MarketInterface
	socket  tradingviewsocket.SocketInterface

	fetchError        error
	failedAPIRequests uint
}

// Run ...
func (s *Manager) Run() {
	s.markets = s.GetMarkets()

	s.initMarkets()
	s.initSocket()

	go s.dailyReset()
	go s.checkSessionDisconnectedError()
	go s.fetchDataLoop()
	go s.panicIfTooManyAPIFails()
}

func (s *Manager) initSocket() {
	s.Logger.Log("Initializing the socket ...")
	tradingviewsocket, err := tradingviewsocket.Connect(
		s.OnReceiveMarketData,
		s.onSocketError,
	)
	if err != nil {
		panic("Error while initializing the trading view socket -> " + err.Error())
	}

	for _, market := range s.markets {
		s.Logger.Log("Adding market to socket ... " + market.GetMarketData().SocketName)
		err = tradingviewsocket.AddSymbol(market.GetMarketData().SocketName)
		if err != nil {
			panic("Error while adding the symbol -> " + err.Error())
		}
	}

	s.socket = tradingviewsocket
}

func (s *Manager) OnReceiveMarketData(marketName string, data *tradingviewsocket.QuoteData) {
	s.Logger.Log("Received data -> " + marketName + " -> " + utils.GetStringRepresentation(data))

	for _, market := range s.markets {
		if marketName != market.GetMarketData().SocketName {
			continue
		}
		if market.GetCurrentBrokerQuote() != nil {
			go market.OnReceiveMarketData(data)
		}
	}
}

func (s *Manager) onSocketError(err error, context string) {
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

func (s *Manager) dailyReset() {
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

func (s *Manager) fetchDataLoop() {
	for {
		var waitingGroup sync.WaitGroup
		var fetchFuncs []func()

		for _, market := range s.markets {
			if !utils.IsNowWithinTradingHours(market.GetMarketData()) {
				continue
			}

			fetchFuncs = append(fetchFuncs,
				func(marketName string) func() {
					return func() {
						quote := s.fetch(func() (interface{}, error) {
							defer waitingGroup.Done()
							s.Logger.Log("Fetching quote ...")
							return s.API.GetQuote(marketName)
						}).(*api.Quote)
						s.Logger.Log("Quote is -> " + utils.GetStringRepresentation(quote))
						if quote != nil {
							for _, market := range s.markets {
								if market.GetMarketData().BrokerAPIName != marketName {
									continue
								}
								market.SetCurrentBrokerQuote(quote)
							}
						}
					}
				}(market.GetMarketData().BrokerAPIName),
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

func (s *Manager) fetch(fetchFunc func() (interface{}, error)) (result interface{}) {
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

func (s *Manager) checkSessionDisconnectedError() {
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

func (s *Manager) panicIfTooManyAPIFails() {
	for {
		// todo: adjust this
		if s.failedAPIRequests >= 500 {
			panic("There is something wrong with the API - Check logs - Stopping bot")
		}
		s.failedAPIRequests = 0
		time.Sleep(1 * time.Minute)
	}
}

func (s *Manager) initMarkets() {
	for _, market := range s.markets {
		s.Logger.Log("Initializing market " + market.GetMarketData().BrokerAPIName)
		go market.Initialize()
	}
}
