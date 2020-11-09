package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// Strategy ...
type Strategy struct {
	API    api.Interface
	Logger logger.Interface

	previousExecutionTime   time.Time
	failedAPIRequestsInARow int
}

// Execute ...
func (s *Strategy) Execute() {
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

	quote, _, positions, state := s.fetchData()

	if quote == nil {
		return
	}

	fmt.Printf("\n%#v\n", quote)
	fmt.Printf("\n%#v\n", positions)
	fmt.Printf("\n%#v\n", state)

	/***
		When creating an order, I need to save the 3 created orders somewhere (the limit/stop order, it's sl and it's tp)
		The SL and the TP will have the parentID of the main one. The main one will have the parentID null
		All 3 orders will have the status "working".

		When modifying an order that hasn't been filled yet, I can use the ID of the main order to change it's sl, tp, or it's limit/stop price.
		When modifying the sl/tp of a position, I need to use the ID of the sl/tp order.
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
	if s.failedAPIRequestsInARow == 100 {
		panic("There is something wrong when fetching the data")
	}

	var waitingGroup sync.WaitGroup
	waitingGroup.Add(4)
	go func() {
		defer waitingGroup.Done()
		quote = s.getQuote("GER30")
	}()

	go func() {
		defer waitingGroup.Done()
		orders = s.getOrders()
	}()

	go func() {
		defer waitingGroup.Done()
		positions = s.getPositions()
	}()

	go func() {
		defer waitingGroup.Done()
		state = s.getState()
	}()

	waitingGroup.Wait()
	s.failedAPIRequestsInARow = 0
	return
}

func (s *Strategy) getQuote(symbol string) (quote *api.Quote) {
	quote, err := s.API.GetQuote(symbol)
	if err != nil {
		s.handleFetchError(err)
		return
	}

	return
}

func (s *Strategy) getOrders() (orders []*api.Order) {
	orders, err := s.API.GetOrders()
	if err != nil {
		s.handleFetchError(err)
		return
	}

	return
}

func (s *Strategy) getPositions() (positions []*api.Position) {
	positions, err := s.API.GetPositions()
	if err != nil {
		s.handleFetchError(err)
		return
	}

	return
}

func (s *Strategy) getState() (state *api.State) {
	state, err := s.API.GetState()
	if err != nil {
		s.handleFetchError(err)
		return
	}

	return
}

func (s *Strategy) handleFetchError(err error) {
	s.failedAPIRequestsInARow++
	s.Logger.Log("Error when fetching - Fails in a row -> " + utils.IntToString(int64(s.failedAPIRequestsInARow)))
	s.Logger.Log("Error was " + err.Error())
}
