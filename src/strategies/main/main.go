package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/logger"
	"strconv"
	"time"
)

// Strategy ...
type Strategy struct {
	API    api.Interface
	Logger logger.Interface

	previousExecutionTime time.Time
	failedAPIRequests     int

	quote     *api.Quote
	orders    []*api.Order
	positions []*api.Position
	state     *api.State
	candles   []*Candle
}

// Execute ...
func (s *Strategy) Execute() {
	go s.panicIfTooManyAPIFails()

	s.candles = []*Candle{&Candle{}}
	go func() {
		for {
			s.quote = s.fetch(func() (interface{}, error) {
				return s.API.GetQuote("GER30")
			}).(*api.Quote)
			if s.quote != nil {
				s.onReceiveMarketData()
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	go func() {
		for {
			s.fetchData()
			time.Sleep(2 * time.Second)
		}
	}()
}

func (s *Strategy) onReceiveMarketData() {
	now := time.Now()

	currentHour, previousHour := s.getCurrentAndPreviousHour(now, s.previousExecutionTime)
	if currentHour == 2 && previousHour == 1 {
		s.candles = nil
		s.candles = []*Candle{&Candle{}}
		s.Logger.ResetLogs()

		s.Logger.Log("Refreshing access token by calling API.Login")
		s.login(120, 30*time.Second)
	}

	if currentHour < 6 || currentHour > 21 {
		s.Logger.Log("Doing nothing - Now it's not the time.")
		return
	}

	s.updateCandles(now)
	s.previousExecutionTime = now
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
