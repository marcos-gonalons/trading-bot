package manager

import (
	"TradingBot/src/markets"
	"TradingBot/src/services"
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/utils"
	"net"
	"strings"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

const DailyResetHour = 2

type Manager struct {
	ServicesContainer *services.Container

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
	s.ServicesContainer.Logger.Log("Initializing the socket ...")
	tradingviewsocket, err := tradingviewsocket.Connect(
		s.OnReceiveMarketData,
		s.onSocketError,
	)
	if err != nil {
		panic("Error while initializing the trading view socket -> " + err.Error())
	}

	for _, market := range s.markets {
		s.ServicesContainer.Logger.Log("Adding market to socket ... " + market.GetMarketData().SocketName)
		err = tradingviewsocket.AddSymbol(market.GetMarketData().SocketName)
		if err != nil {
			panic("Error while adding the symbol -> " + err.Error())
		}
	}

	s.socket = tradingviewsocket
}

func (s *Manager) OnReceiveMarketData(marketName string, data *tradingviewsocket.QuoteData) {
	s.ServicesContainer.Logger.Log("Received data -> " + marketName + " -> " + utils.GetStringRepresentation(data))

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
	s.ServicesContainer.Logger.Log("Socket error -> " + err.Error())
	s.ServicesContainer.Logger.Log("Context -> " + context)
	err = s.socket.Close()
	if err != nil {
		s.ServicesContainer.Logger.Error("Error when closing the socket -> " + err.Error())
		if !strings.Contains(err.Error(), "use of closed network connection") {
			return
		}
	}

	s.ServicesContainer.Logger.Log("Initializing the socket again ... ")
	s.initSocket()
}

func (s *Manager) dailyReset() {
	for {
		currentHour, _ := utils.GetCurrentTimeHourAndMinutes()

		if currentHour == DailyResetHour {
			s.ServicesContainer.Logger.ResetLogs()

			err := s.socket.Close()
			if err != nil {
				s.ServicesContainer.Logger.Error("Error when restarting the socket -> " + err.Error())
			}

			for _, market := range s.markets {
				go market.DailyReset()
			}

			s.ServicesContainer.Logger.Log("Refreshing access token by calling API.Login")
			s.ServicesContainer.APIRetryFacade.Login(retryFacade.RetryParams{
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
							s.ServicesContainer.Logger.Log("Fetching quote ...")
							return s.ServicesContainer.API.GetQuote(marketName)
						}).(*api.Quote)
						s.ServicesContainer.Logger.Log("Quote is -> " + utils.GetStringRepresentation(quote))
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
						s.ServicesContainer.Logger.Log("Fetching orders ...")
						return s.ServicesContainer.API.GetOrders()
					}).([]*api.Order)
					s.ServicesContainer.Logger.Log("Orders is -> " + utils.GetStringRepresentation(orders))
					var o []*api.Order
					if orders != nil {
						o = orders
					}
					s.ServicesContainer.APIData.SetOrders(o)
				},
				func() {
					positions := s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						s.ServicesContainer.Logger.Log("Fetching positions ...")
						return s.ServicesContainer.API.GetPositions()
					}).([]*api.Position)
					s.ServicesContainer.Logger.Log("Positions is -> " + utils.GetStringRepresentation(positions))
					var p []*api.Position
					if positions != nil {
						p = positions
					}
					s.ServicesContainer.APIData.SetPositions(p)
				},
				func() {
					state := s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						s.ServicesContainer.Logger.Log("Fetching state ...")
						return s.ServicesContainer.API.GetState()
					}).(*api.State)
					s.ServicesContainer.Logger.Log("State is -> " + utils.GetStringRepresentation(state))
					if state != nil {
						s.ServicesContainer.APIData.SetState(state)
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
	s.ServicesContainer.Logger.Error("Error while fetching data -> " + err.Error())
	s.failedAPIRequests++

	if err, isNetError := err.(net.Error); isNetError && err.Timeout() {
		currentTimeout := s.ServicesContainer.API.GetTimeout()
		if currentTimeout < 60*time.Second {
			s.ServicesContainer.Logger.Log("Increasing timeout ...")
			s.ServicesContainer.API.SetTimeout(currentTimeout + 10*time.Second)
		} else {
			s.ServicesContainer.Logger.Error("The API has a timeout problem")
		}
	} else {
		s.ServicesContainer.Logger.Log("Setting default timeout of 10 seconds ...")
		s.ServicesContainer.API.SetTimeout(10 * time.Second)
	}
	return
}

func (s *Manager) checkSessionDisconnectedError() {
	for {
		if s.ServicesContainer.API.IsSessionDisconnectedError(s.fetchError) {
			s.ServicesContainer.Logger.Log("Session is disconnected. Loggin in again ... ")
			s.ServicesContainer.APIRetryFacade.Login(retryFacade.RetryParams{
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
		s.ServicesContainer.Logger.Log("Initializing market " + market.GetMarketData().BrokerAPIName)
		go market.Initialize()
	}
}
