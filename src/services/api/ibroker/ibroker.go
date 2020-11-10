package ibroker

import (
	closeorder "TradingBot/src/services/api/ibroker/closeOrder"
	closeposition "TradingBot/src/services/api/ibroker/closePosition"
	"TradingBot/src/services/api/ibroker/createorder"
	getorders "TradingBot/src/services/api/ibroker/getOrders"
	getpositions "TradingBot/src/services/api/ibroker/getPositions"
	getstate "TradingBot/src/services/api/ibroker/getState"
	"TradingBot/src/services/api/ibroker/getquote"
	"TradingBot/src/services/api/ibroker/login"
	modifyorder "TradingBot/src/services/api/ibroker/modifyOrder"
	modifyposition "TradingBot/src/services/api/ibroker/modifyPosition"
	"TradingBot/src/services/logger"

	"encoding/json"
	"net/http"
	"time"

	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
)

// API ...
type API struct {
	httpclient httpclient.Interface
	logger     logger.Interface

	accessToken *api.AccessToken
	credentials *api.Credentials
	orders      []*api.Order
	positions   []*api.Position
}

// Login ...
func (s *API) Login() (accessToken *api.AccessToken, err error) {
	defer func() {
		s.logAPIResult(accessToken, err, logger.LoginRequest)
	}()

	url := s.getURL("authorize")
	accessToken, err = login.Request(
		url,
		s.credentials,
		s.httpclient,
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func() error {
			return s.doOptionsRequest(url, http.MethodPost, logger.LoginRequest)
		},
	)

	s.accessToken = accessToken
	return
}

// GetQuote ...
func (s *API) GetQuote(symbol string) (quote *api.Quote, err error) {
	defer func() {
		s.logAPIResult(quote, err, logger.GetQuoteRequest)
	}()

	url := s.getURL("quotes")
	quote, err = getquote.Request(
		url,
		s.httpclient,
		s.accessToken.Token,
		s.credentials.AccountID,
		symbol,
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func() error {
			return s.doOptionsRequest(url, http.MethodGet, logger.GetQuoteRequest)
		},
	)

	return
}

// CreateOrder ...
func (s *API) CreateOrder(order *api.Order) (err error) {
	defer func() {
		s.logAPIResult("", err, logger.CreateOrderRequest)
	}()

	url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/orders"
	err = createorder.Request(
		url,
		s.httpclient,
		s.accessToken.Token,
		order,
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func() error {
			return s.doOptionsRequest(url, http.MethodPost, logger.CreateOrderRequest)
		},
	)

	return
}

// GetOrders ...
func (s *API) GetOrders() (orders []*api.Order, err error) {
	defer func() {
		s.logAPIResult(orders, err, logger.GetOrdersRequest)
	}()

	url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/orders"
	orders, err = getorders.Request(
		url,
		s.httpclient,
		s.accessToken.Token,
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func() error {
			return s.doOptionsRequest(url, http.MethodGet, logger.GetOrdersRequest)
		},
	)

	s.orders = orders
	return
}

// ModifyOrder ...
func (s *API) ModifyOrder(order *api.Order) (err error) {
	defer func() {
		s.logAPIResult("", err, logger.ModifyOrderRequest)
	}()

	url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/orders/" + order.ID
	err = modifyorder.Request(
		url,
		s.httpclient,
		s.accessToken.Token,
		order,
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func() error {
			return s.doOptionsRequest(url, http.MethodPut, logger.ModifyOrderRequest)
		},
	)

	return
}

// ClosePosition ...
func (s *API) ClosePosition(symbol string) (err error) {
	defer func() {
		s.logAPIResult("", err, logger.ClosePositionRequest)
	}()

	url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/positions/" + symbol
	err = closeposition.Request(
		url,
		s.httpclient,
		s.accessToken.Token,
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func() error {
			return s.doOptionsRequest(url, http.MethodDelete, logger.ClosePositionRequest)
		},
	)

	return
}

// CloseOrder ...
func (s *API) CloseOrder(orderID string) (err error) {
	defer func() {
		s.logAPIResult("", err, logger.CloseOrderRequest)
	}()

	url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/orders/" + orderID
	err = closeorder.Request(
		url,
		s.httpclient,
		s.accessToken.Token,
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func() error {
			return s.doOptionsRequest(url, http.MethodDelete, logger.CloseOrderRequest)
		},
	)

	return
}

// GetPositions ...
func (s *API) GetPositions() (positions []*api.Position, err error) {
	defer func() {
		s.logAPIResult(positions, err, logger.GetPositionsRequest)
	}()

	url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/positions"
	positions, err = getpositions.Request(
		url,
		s.httpclient,
		s.accessToken.Token,
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func() error {
			return s.doOptionsRequest(url, http.MethodGet, logger.GetPositionsRequest)
		},
	)

	s.positions = positions
	return
}

// CloseEverything ...
func (s *API) CloseEverything() (err error) {
	if len(s.positions) > 0 {
		for _, position := range s.positions {
			s.ClosePosition(position.Instrument)
		}
	}
	if len(s.orders) > 0 {
		workingOrders := getWorkingOrders(s.orders)
		for _, order := range workingOrders {
			if order.ParentID == nil {
				s.CloseOrder(order.ID)
			}
		}

		orders, err := s.GetOrders()
		if err != nil {
			return err
		}

		workingOrders = getWorkingOrders(orders)
		for _, order := range workingOrders {
			s.CloseOrder(order.ID)
		}
	}
	return
}

// GetState ...
func (s *API) GetState() (state *api.State, err error) {
	defer func() {
		s.logAPIResult(state, err, logger.GetStateRequest)
	}()

	url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/state?locale=en"
	state, err = getstate.Request(
		url,
		s.httpclient,
		s.accessToken.Token,
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func() error {
			return s.doOptionsRequest(url, http.MethodGet, logger.GetStateRequest)
		},
	)

	return
}

// ModifyPosition ...
func (s *API) ModifyPosition(symbol string, takeProfit *string, stopLoss *string) (err error) {
	defer func() {
		s.logAPIResult("", err, logger.ModifyPositionRequest)
	}()

	url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/positions/" + symbol
	err = modifyposition.Request(
		url,
		s.httpclient,
		s.accessToken.Token,
		takeProfit,
		stopLoss,
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func() error {
			return s.doOptionsRequest(url, http.MethodPut, logger.ModifyPositionRequest)
		},
	)

	return
}

func (s *API) getURL(endpoint string) string {
	return "https://www.ibroker.es/tradingview/api/" + endpoint
}

func (s *API) setHeaders(rq *http.Request, isOptionsRequest bool, method string) {
	rq.Header.Set("Host", "www.ibroker.es")
	rq.Header.Set("Connection", "keep-alive")
	rq.Header.Set("Pragma", "no-cache")
	rq.Header.Set("Cache-Control", "no-cache")
	if isOptionsRequest {
		rq.Header.Set("Accept", "*/*")
		rq.Header.Set("Access-Control-Request-Method", method)
		rq.Header.Set("Access-Control-Request-Headers", "authorization")
	} else {
		rq.Header.Set("Accept", "application/json")
	}
	rq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36")
	rq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rq.Header.Set("Origin", "https://www.tradingview.com")
	rq.Header.Set("Sec-Fetch-Site", "cross-site")
	rq.Header.Set("Sec-Fetch-Mode", "cors")
	rq.Header.Set("Sec-Fetch-Dest", "empty")
	rq.Header.Set("Referer", "https://www.tradingview.com/")
	rq.Header.Set("Accept-Encoding", "gzip, deflate, br")
	rq.Header.Set("Accept-Language", "en-US,en;q=0.9,es;q=0.8")
}

func (s *API) doOptionsRequest(url string, method string, logType logger.LogType) (err error) {
	// The OPTIONS request is, of course, not really needed outside a browser's context.
	// But since we want to simulate we are in the browser, we send the options request anyway.
	rq, err := http.NewRequest(
		http.MethodOptions,
		url,
		nil,
	)
	if err != nil {
		return
	}

	s.setHeaders(rq, true, method)
	_, err = s.httpclient.Do(rq, logType)
	if err != nil {
		return
	}

	return
}

func (s *API) setCredentials(credentials *api.Credentials) {
	s.credentials = credentials
}

func (s *API) logAPIResult(response interface{}, err error, logType logger.LogType) {
	if err != nil {
		s.logger.Log("ERROR -> "+err.Error(), logType)
	} else {
		str, err := json.Marshal(response)
		if err != nil {
			s.logger.Log("ERROR -> "+err.Error(), logType)
		} else {
			s.logger.Log("RESULT -> "+string(str), logType)
		}
	}
}

func getWorkingOrders(orders []*api.Order) []*api.Order {
	var workingOrders []*api.Order
	for _, order := range orders {
		if order.Status == "working" {
			workingOrders = append(workingOrders, order)
		}
	}
	return workingOrders
}

// CreateAPIServiceInstance ...
func CreateAPIServiceInstance(credentials *api.Credentials) api.Interface {
	httpclient := &httpclient.Service{
		Logger: logger.GetInstance(),
	}
	httpclient.SetTimeout(time.Second * 10)

	instance := &API{
		httpclient: httpclient,
		logger:     logger.GetInstance(),
	}
	instance.setCredentials(credentials)

	return instance
}
