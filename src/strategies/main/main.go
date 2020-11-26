package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
	"math"
	"net"
	"strconv"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper"
)

// Strategy ...
type Strategy struct {
	API    api.Interface
	Logger logger.Interface

	currentExecutionTime  time.Time
	previousExecutionTime time.Time
	failedAPIRequests     int

	currentBrokerQuote *api.Quote
	orders             []*api.Order
	pendingOrder       *api.Order
	positions          []*api.Position
	state              *api.State

	candles []*Candle

	csvFileName string
	csvFileMtx  sync.Mutex

	socket        tradingviewsocket.SocketInterface
	lastVolume    float64
	lastBid       *float64
	lastAsk       *float64
	spreads       []float64
	averageSpread float64

	creatingOrderTimestamp     int64
	modifyingOrderTimestamp    int64
	modifyingPositionTimestamp int64
	closingOrdersTimestamp     int64

	fetchError error
}

// Execute ...
func (s *Strategy) Execute() {
	s.initSocket()
	s.initCandles()

	go s.resetAtTwoAm()
	go s.panicIfTooManyAPIFailsOrSocketErrors()
	go s.checkSessionDisconnectedError()
	go s.fetchDataLoop()
}

func (s *Strategy) resetAtTwoAm() {
	for {
		currentHour, _ := strconv.Atoi(time.Now().Format("15"))

		if currentHour == 2 {
			s.Logger.ResetLogs()

			err := s.socket.Close()
			if err != nil {
				s.Logger.Error("Error when restarting the socket -> " + err.Error())
			}
			s.initCandles()
			s.pendingOrder = nil

			s.Logger.Log("Refreshing access token by calling API.Login")
			s.login(120, 30*time.Second)
		}

		time.Sleep(60 * time.Minute)
	}
}

func (s *Strategy) initSocket() {
	s.Logger.Log("Initializing the socket ...")
	tradingviewsocket, err := tradingviewsocket.Connect(
		s.onReceiveMarketData,
		s.onSocketError,
	)
	if err != nil {
		panic("Error while initializing the trading view socket -> " + err.Error())
	}

	err = tradingviewsocket.AddSymbol("FX:GER30")
	if err != nil {
		panic("Error while adding the symbol -> " + err.Error())
	}

	s.socket = tradingviewsocket
}

func (s *Strategy) fetchDataLoop() {
	for {
		var waitingGroup sync.WaitGroup
		currentHour, _ := strconv.Atoi(s.currentExecutionTime.Format("15"))
		if currentHour >= 6 && currentHour <= 21 {
			fetchFuncs := []func(){
				func() {
					s.currentBrokerQuote = s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						return s.API.GetQuote(ibroker.GER30SymbolName)
					}).(*api.Quote)
				},
				func() {
					s.orders = s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						return s.API.GetOrders()
					}).([]*api.Order)
				},
				func() {
					s.positions = s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						return s.API.GetPositions()
					}).([]*api.Position)
				},
				func() {
					s.state = s.fetch(func() (interface{}, error) {
						defer waitingGroup.Done()
						return s.API.GetState()
					}).(*api.State)
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

func (s *Strategy) checkSessionDisconnectedError() {
	for {
		if s.fetchError != nil && s.fetchError.Error() == "Api error -> Your session is disconnected. Please login again to initialize a new valid session." {
			s.Logger.Log("Session is disconnected. Loggin in again ... ")
			s.login(0, 0)
		}
		s.fetchError = nil
		time.Sleep(5 * time.Second)
	}
}

func (s *Strategy) onReceiveMarketData(symbol string, data *tradingviewsocket.QuoteData) {
	s.Logger.Log("Received data -> " + symbol + " -> " + utils.GetStringRepresentation(data))

	s.currentExecutionTime = time.Now()
	defer func() {
		s.previousExecutionTime = s.currentExecutionTime

		if data.Volume != nil {
			s.lastVolume = *data.Volume
		}
		if data.Bid != nil {
			s.lastBid = data.Bid
		}
		if data.Ask != nil {
			s.lastAsk = data.Ask
		}
	}()
	s.updateCandles(data)

	go s.updateAverageSpread()
	if len(s.candles) == 0 {
		return
	}
	go s.breakoutAnticipationStrategy()
}

func (s *Strategy) updateAverageSpread() {
	if s.lastAsk == nil || s.lastBid == nil {
		return
	}

	if len(s.spreads) == 1500 {
		s.spreads = s.spreads[1:]
	}

	s.spreads = append(s.spreads, math.Abs(*s.lastAsk-*s.lastBid))

	var sum float64
	for _, spread := range s.spreads {
		sum += spread
	}

	s.averageSpread = sum / float64(len(s.spreads))
}

func (s *Strategy) getCurrentAndPreviousHour() (int, int) {
	currentHour, _ := strconv.Atoi(s.currentExecutionTime.Format("15"))
	previousHour, _ := strconv.Atoi(s.previousExecutionTime.Format("15"))
	return currentHour, previousHour
}

func (s *Strategy) fetch(fetchFunc func() (interface{}, error)) (result interface{}) {
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

func (s *Strategy) panicIfTooManyAPIFailsOrSocketErrors() {
	for {
		if s.failedAPIRequests >= 50 {
			panic("There is something wrong with the API - Check logs - Stopping bot")
		}
		s.failedAPIRequests = 0
		time.Sleep(1 * time.Minute)
	}
}

func (s *Strategy) onSocketError(err error) {
	s.Logger.Log("Socket error -> " + err.Error())
	s.socket.Close()
	s.initSocket()
}
