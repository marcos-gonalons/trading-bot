package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/logger"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper"
)

// Strategy ...
type Strategy struct {
	API    api.Interface
	Logger logger.Interface

	previousExecutionTime time.Time
	failedAPIRequests     int

	orders    []*api.Order
	positions []*api.Position
	state     *api.State
	candles   []*Candle

	csvFileName string
	csvFileMtx  sync.Mutex

	socket        tradingviewsocket.SocketInterface
	currentVolume float64
}

// Execute ...
func (s *Strategy) Execute() {
	go s.panicIfTooManyAPIFails()

	s.initSocket()
	s.initCandles()
	go func() {
		for {
			now := time.Now()
			currentHour, _ := strconv.Atoi(now.Format("15"))
			if currentHour >= 6 && currentHour <= 21 {
				s.fetchData()
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func (s *Strategy) onReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.Logger.Log("Received data -> " + symbol + " -> " + getStringRepresentation(data))

	now := time.Now()

	if data.Volume != nil {
		s.currentVolume = *data.Volume
	}

	currentHour, previousHour := s.getCurrentAndPreviousHour(now, s.previousExecutionTime)
	if currentHour == 2 && previousHour == 1 {
		s.initCandles()
		s.Logger.ResetLogs()

		s.Logger.Log("Refreshing access token by calling API.Login")
		s.login(120, 30*time.Second)
	}

	s.updateCandles(now, data)

	s.previousExecutionTime = now
	if currentHour < 6 || currentHour > 21 {
		s.Logger.Log("Doing nothing - Now it's not the time.")
		return
	}
}

func (s *Strategy) getCurrentAndPreviousHour(
	now time.Time,
	previous time.Time,
) (int, int) {
	currentHour, _ := strconv.Atoi(now.Format("15"))
	previousHour, _ := strconv.Atoi(previous.Format("15"))
	return currentHour, previousHour
}

func (s *Strategy) login(maxRetries uint, timeBetweenRetries time.Duration) {
	_, err := s.API.Login()
	if err == nil {
		return
	}

	go func() {
		var err error
		var retriesAmount uint = 0
		for {
			if retriesAmount == maxRetries {
				panic("Too many failed login attempts. Last error was " + err.Error())
			}
			_, err = s.API.Login()

			if err == nil {
				break
			}

			retriesAmount++
			time.Sleep(timeBetweenRetries)
		}
	}()
}

func (s *Strategy) fetchData() {
	fetchFuncs := []func(){
		func() {
			s.orders = s.fetch(func() (interface{}, error) {
				return s.API.GetOrders()
			}).([]*api.Order)
		},
		func() {
			s.positions = s.fetch(func() (interface{}, error) {
				return s.API.GetPositions()
			}).([]*api.Position)
		},
		func() {
			s.state = s.fetch(func() (interface{}, error) {
				return s.API.GetState()
			}).(*api.State)
		},
	}

	for _, fetchFunc := range fetchFuncs {
		go fetchFunc()
	}
	return
}

func (s *Strategy) fetch(fetchFunc func() (interface{}, error)) (result interface{}) {
	result, err := fetchFunc()

	if err != nil {
		s.failedAPIRequests++
		s.Logger.Log("Error while fetching data -> " + err.Error())
		return
	}

	return
}

func (s *Strategy) panicIfTooManyAPIFails() {
	for {
		if s.failedAPIRequests >= 15 {
			panic("There is something wrong with the API - Check logs - Stopping bot")
		}
		s.failedAPIRequests = 0
		time.Sleep(1 * time.Minute)
	}
}

func (s *Strategy) onSocketError(err error) {
	// TODO: Probably reset the socket connection
	panic("Socket error " + err.Error())
}

func (s *Strategy) initSocket() {
	/**
		fx va 1 pip por detras
		si fx:ger30 dice 13149.4, en ibroker es 13150.5

		leer con unauthorized de fx:ger30
		y tener en cuenta que va 1 por detras

		El volumen se resetea cada dia a las 23:00 hora de espanya (al menos en eurusd)
		Y cuando se recibe el volumen se recibe el volumen acumulado desde el reseteo hasta ese momento.
	**/

	tradingviewsocket, err := tradingviewsocket.Connect(
		s.onReceiveMarketData,
		s.onSocketError,
	)
	if err != nil {
		panic("Error while initializing the trading view socket -> " + err.Error())
	}

	err = tradingviewsocket.AddSymbol("FX:EURUSD")
	if err != nil {
		panic("Error while adding the symbol -> " + err.Error())
	}

	s.socket = tradingviewsocket
}

func getStringRepresentation(data *tradingviewsocket.QuoteData) string {
	str, _ := json.Marshal(data)
	return string(str)
}
