package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/logger"
	"strconv"
	"sync"
	"time"
)

// Strategy ...
type Strategy struct {
	API    api.Interface
	Logger logger.Interface

	previousExecutionTime time.Time
	failedAPIRequests     int
}

// Execute ...
func (s *Strategy) Execute() {
	go s.panicIfTooManyAPIFails()
	for {
		s.execute()
		// Why 1.66666 seconds?
		// Tradingview sends the get requests every 1.66666 seconds, so we should do the same.
		time.Sleep(1666666 * time.Microsecond)
	}
}

func (s *Strategy) execute() {
	now := time.Now()

	currentHour, previousHour := s.getCurrentAndPreviousHour(now, s.previousExecutionTime)
	if currentHour == 2 && previousHour == 1 {
		s.Logger.ResetLogs()

		s.Logger.Log("Refreshing access token by calling API.Login")
		s.login(120, 30*time.Second)
	}

	if currentHour < 6 || currentHour > 21 {
		s.Logger.Log("Doing nothing - Now it's not the time.")
		return
	}

	quote, _, _, _ := s.fetchData()

	if quote == nil {
		return
	}

	/***
		When creating an order, I need to save the 3 created orders somewhere (the limit/stop order, it's sl and it's tp)
		The SL and the TP will have the parentID of the main one. The main one will have the parentID null
		All 3 orders will have the status "working".

		When modifying an order that hasn't been filled yet, I can use the ID of the main order to change it's sl, tp, or it's limit/stop price.
		When modifying the sl/tp of a position, I need to use the ID of the sl/tp order.



		Take into consideration
		Let's say the bot dies, for whatever reason, at 15:00pm
		I revive him at 15:05
		It will have lost all the candles[]

		To mitigate this
		As I add a candle to the candles[]
		Save the candles to the csv file
		When booting the bot; initialize the candles array with those in the csv file


		When booting {
			if !csv file, create the csv file
			else, load candles[] from the file
		}
	***/

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

func (s *Strategy) fetchData() (
	quote *api.Quote,
	orders []*api.Order,
	positions []*api.Position,
	state *api.State,
) {
	var waitingGroup sync.WaitGroup
	fetchFuncs := []func(){
		func() {
			defer waitingGroup.Done()
			quote = s.fetch(func() (interface{}, error) {
				return s.API.GetQuote("GER30")
			}).(*api.Quote)
		},
		func() {
			defer waitingGroup.Done()
			orders = s.fetch(func() (interface{}, error) {
				return s.API.GetOrders()
			}).([]*api.Order)
		},
		func() {
			defer waitingGroup.Done()
			positions = s.fetch(func() (interface{}, error) {
				return s.API.GetPositions()
			}).([]*api.Position)
		},
		func() {
			defer waitingGroup.Done()
			state = s.fetch(func() (interface{}, error) {
				return s.API.GetState()
			}).(*api.State)
		},
	}

	waitingGroup.Add(len(fetchFuncs))
	for _, fetchFunc := range fetchFuncs {
		go fetchFunc()
	}
	waitingGroup.Wait()
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
