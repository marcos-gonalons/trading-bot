package ibroker

import (
	closeorder "TradingBot/src/services/api/ibroker/closeOrder"
	closeposition "TradingBot/src/services/api/ibroker/closePosition"
	createorder "TradingBot/src/services/api/ibroker/createOrder"
	getorders "TradingBot/src/services/api/ibroker/getOrders"
	getpositions "TradingBot/src/services/api/ibroker/getPositions"
	getquote "TradingBot/src/services/api/ibroker/getQuote"
	getstate "TradingBot/src/services/api/ibroker/getState"
	"TradingBot/src/services/api/ibroker/login"
	modifyorder "TradingBot/src/services/api/ibroker/modifyOrder"
	modifyposition "TradingBot/src/services/api/ibroker/modifyPosition"
	"TradingBot/src/services/logger"
	loggerTypes "TradingBot/src/services/logger/types"
	"errors"

	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/types"
)

// API ...
type API struct {
	httpclient httpclient.Interface
	logger     logger.Interface
	url        string

	timeout     time.Duration
	accessToken *api.AccessToken
	credentials *api.Credentials
	orders      []*api.Order
	positions   []*api.Position
	state       *api.State
}

// Login ...
func (s *API) Login() (accessToken *api.AccessToken, err error) {
	returnValue, err := s.apiCall(
		loggerTypes.LoginRequest,
		func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (r interface{}, e error) {
			url := s.getURL("authorize")
			return login.Request(
				url,
				s.httpclient,
				setHeaders,
				optionsRequest,
				&login.RequestParameters{
					Credentials: s.credentials,
				},
			)
		},
	)
	accessToken = returnValue.(*api.AccessToken)
	s.accessToken = accessToken
	return
}

// GetQuote ...
func (s *API) GetQuote(marketName string) (quote *api.Quote, err error) {
	returnValue, err := s.apiCall(
		loggerTypes.GetQuoteRequest,
		func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (r interface{}, e error) {
			url := s.getURL("quotes")
			return getquote.Request(
				url,
				s.httpclient,
				setHeaders,
				optionsRequest,
				&getquote.RequestParameters{
					AccessToken: s.accessToken.Token,
					AccountID:   s.credentials.AccountID,
					Symbol:      marketName,
				},
			)
		},
	)
	quote = returnValue.(*api.Quote)
	return
}

// CreateOrder ...
func (s *API) CreateOrder(order *api.Order) (err error) {
	_, err = s.apiCall(
		loggerTypes.CreateOrderRequest,
		func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (r interface{}, e error) {
			url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/orders"
			return createorder.Request(
				url,
				s.httpclient,
				setHeaders,
				optionsRequest,
				&createorder.RequestParameters{
					AccessToken: s.accessToken.Token,
					Order:       order,
				},
			)
		},
	)
	return
}

// GetOrders ...
func (s *API) GetOrders() (orders []*api.Order, err error) {
	returnValue, err := s.apiCall(
		loggerTypes.GetOrdersRequest,
		func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (r interface{}, e error) {
			url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/orders"
			return getorders.Request(
				url,
				s.httpclient,
				setHeaders,
				optionsRequest,
				&getorders.RequestParameters{
					AccessToken: s.accessToken.Token,
				},
			)
		},
	)
	orders = returnValue.([]*api.Order)
	s.orders = orders
	return
}

// SetOrders ...
func (s *API) SetOrders(orders []*api.Order) {
	s.orders = orders
}

// ModifyOrder ...
func (s *API) ModifyOrder(order *api.Order) (err error) {
	_, err = s.apiCall(
		loggerTypes.ModifyOrderRequest,
		func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (r interface{}, e error) {
			url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/orders/" + order.ID
			return modifyorder.Request(
				url,
				s.httpclient,
				setHeaders,
				optionsRequest,
				&modifyorder.RequestParameters{
					AccessToken:  s.accessToken.Token,
					Order:        order,
					IsLimitOrder: s.IsLimitOrder,
					IsStopOrder:  s.IsStopOrder,
				},
			)
		},
	)
	return
}

// ClosePosition ...
func (s *API) ClosePosition(marketName string) (err error) {
	slOrder, tpOrder := s.GetBracketOrders(marketName)
	if slOrder != nil {
		err = s.CloseOrder(slOrder.ID)
	}
	if tpOrder != nil {
		err = s.CloseOrder(tpOrder.ID)
	}
	if err != nil {
		return
	}

	_, err = s.apiCall(
		loggerTypes.ClosePositionRequest,
		func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (r interface{}, e error) {
			url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/positions/" + url.QueryEscape(marketName)
			return closeposition.Request(
				url,
				s.httpclient,
				setHeaders,
				optionsRequest,
				&closeposition.RequestParameters{
					AccessToken: s.accessToken.Token,
				},
			)
		},
	)
	return
}

// AddTrade ...
func (s *API) AddTrade(
	order *api.Order,
	position *api.Position,
	slippageFunc func(price float64, order *api.Order) float64,
	eurExchangeRate float64,
	lastCandle *types.Candle,
	marketData *types.MarketData,
) {
	// Nothing to do here ...
}

// GetTrades ...
func (s *API) GetTrades() []*api.Trade {
	// Nothing to do here ...
	return nil
}

// SetTrades ...
func (s *API) SetTrades(t []*api.Trade) {
}

// CloseOrder ...
func (s *API) CloseOrder(orderID string) (err error) {
	_, err = s.apiCall(
		loggerTypes.CloseOrderRequest,
		func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (r interface{}, e error) {
			url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/orders/" + orderID
			return closeorder.Request(
				url,
				s.httpclient,
				setHeaders,
				optionsRequest,
				&closeorder.RequestParameters{
					AccessToken: s.accessToken.Token,
				},
			)
		},
	)

	return
}

// GetPositions ...
func (s *API) GetPositions() (positions []*api.Position, err error) {
	returnValue, err := s.apiCall(
		loggerTypes.GetPositionsRequest,
		func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (r interface{}, e error) {
			url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/positions"
			return getpositions.Request(
				url,
				s.httpclient,
				setHeaders,
				optionsRequest,
				&getpositions.RequestParameters{
					AccessToken: s.accessToken.Token,
				},
			)
		},
	)

	positions = returnValue.([]*api.Position)
	s.positions = positions
	return
}

// SetPositions ...
func (s *API) SetPositions(positions []*api.Position) {
	s.positions = positions
}

// GetState ...
func (s *API) GetState() (state *api.State, err error) {
	returnValue, err := s.apiCall(
		loggerTypes.GetStateRequest,
		func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (r interface{}, e error) {
			url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/state?locale=en"
			return getstate.Request(
				url,
				s.httpclient,
				setHeaders,
				optionsRequest,
				&getstate.RequestParameters{
					AccessToken: s.accessToken.Token,
				},
			)
		},
	)

	state = returnValue.(*api.State)
	return
}

// SetState ...
func (s *API) SetState(state *api.State) {
	s.state = state
}

// ModifyPosition ...
func (s *API) ModifyPosition(marketName string, takeProfit *string, stopLoss *string) (err error) {
	_, err = s.apiCall(
		loggerTypes.ModifyPositionRequest,
		func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (r interface{}, e error) {
			url := s.getURL("accounts") + "/" + s.credentials.AccountID + "/positions/" + url.QueryEscape(marketName)
			return modifyposition.Request(
				url,
				s.httpclient,
				setHeaders,
				optionsRequest,
				&modifyposition.RequestParameters{
					AccessToken: s.accessToken.Token,
					TakeProfit:  takeProfit,
					StopLoss:    stopLoss,
				},
			)
		},
	)
	return
}

// GetWorkingOrders ...
func (s *API) GetWorkingOrders(orders []*api.Order) []*api.Order {
	var workingOrders []*api.Order
	for _, order := range orders {
		if order.Status == "working" {
			workingOrders = append(workingOrders, order)
		}
	}
	return workingOrders
}

// CloseAllOrders ...
func (s *API) CloseAllOrders() (err error) {
	if len(s.orders) == 0 {
		return
	}

	// First, close the main order that is not the TP nor the SL
	errString := ""
	workingOrders := s.GetWorkingOrders(s.orders)
	for _, order := range workingOrders {
		if order.ParentID != nil {
			continue
		}
		err = s.CloseOrder(order.ID)
		if err != nil {
			errString += err.Error() + "\n"
		}
	}

	// Then, get the orders again from the broker and close the TP and SL orders
	orders, err := s.GetOrders()
	if err != nil {
		errString += err.Error() + "\n"
		err = errors.New(errString)
		return
	}

	workingOrders = s.GetWorkingOrders(orders)
	for _, order := range workingOrders {
		err = s.CloseOrder(order.ID)
		if err != nil {
			errString += err.Error() + "\n"
		}
	}
	if errString != "" {
		err = errors.New(errString)
	}

	return
}

// CloseAllPositions ...
func (s *API) CloseAllPositions() (err error) {
	if len(s.positions) == 0 {
		return
	}
	errString := ""
	for _, position := range s.positions {
		err = s.ClosePosition(position.Instrument)
		if err != nil {
			errString += err.Error() + "\n"
		}
	}
	if errString != "" {
		err = errors.New(errString)
	}

	return
}

// SetTimeout ...
func (s *API) SetTimeout(t time.Duration) {
	s.timeout = t
	s.httpclient.SetTimeout(t)
}

// GetTimeout ...
func (s *API) GetTimeout() time.Duration {
	return s.timeout
}

func (s *API) GetBracketOrders(marketName string) (
	slOrder *api.Order,
	tpOrder *api.Order,
) {
	for _, order := range s.orders {
		if order.Status != "working" || order.Instrument != marketName {
			continue
		}
		if s.IsLimitOrder(order) {
			tpOrder = order
		}
		if s.IsStopOrder(order) {
			slOrder = order
		}
	}
	return
}

func (s *API) GetWorkingOrderWithBracketOrders(side string, marketName string, orders []*api.Order) []*api.Order {
	var workingOrders []*api.Order

	for _, order := range s.orders {
		if order.Status != "working" || order.Side != side || order.Instrument != marketName || order.ParentID != nil {
			continue
		}

		workingOrders = append(workingOrders, order)
	}

	if len(workingOrders) > 0 {
		for _, order := range s.orders {
			if order.Status != "working" || order.ParentID == nil || *order.ParentID != workingOrders[0].ID {
				continue
			}

			workingOrders = append(workingOrders, order)
		}
	}

	return workingOrders
}

func (s *API) GetSLAndTPOrders(parentID string, orders []*api.Order) (*api.Order, *api.Order) {
	var slOrder *api.Order
	var tpOrder *api.Order
	for _, workingOrder := range orders {
		if workingOrder.ParentID == nil || *workingOrder.ParentID != parentID {
			continue
		}

		if s.IsLimitOrder(workingOrder) {
			tpOrder = workingOrder
		}
		if s.IsStopOrder(workingOrder) {
			slOrder = workingOrder
		}
	}

	return slOrder, tpOrder
}

func (s *API) apiCall(
	logType loggerTypes.LogType,
	call func(setHeaders func(rq *http.Request), optionsRequest func(url string, httpMethod string) error) (interface{}, error),
) (r interface{}, err error) {
	defer func() {
		s.logAPIResult(r, err, logType)
	}()

	r, err = call(
		func(rq *http.Request) {
			s.setHeaders(rq, false, "")
		},
		func(url string, httpMethod string) error {
			return s.optionsRequest(url, httpMethod, logType)
		},
	)

	return
}

func (s *API) getURL(endpoint string) string {
	return s.url + endpoint
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
	rq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")
	rq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rq.Header.Set("Origin", "https://www.tradingview.com")
	rq.Header.Set("Sec-Fetch-Site", "cross-site")
	rq.Header.Set("Sec-Fetch-Mode", "cors")
	rq.Header.Set("Sec-Fetch-Dest", "empty")
	rq.Header.Set("Referer", "https://www.tradingview.com/")
	rq.Header.Set("Accept-Encoding", "gzip, deflate, br")
	rq.Header.Set("Accept-Language", "en-US,en;q=0.9,es;q=0.8")
}

func (s *API) optionsRequest(url string, method string, logType loggerTypes.LogType) (err error) {
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

func (s *API) logAPIResult(response interface{}, err error, logType loggerTypes.LogType) {
	if err != nil {
		s.logger.Error("ERROR -> "+err.Error(), logType)
	} else {
		str, err := json.Marshal(response)
		if err != nil {
			s.logger.Error("ERROR -> "+err.Error(), logType)
		} else {
			s.logger.Log("RESULT -> "+string(str), logType)
		}
	}
}

func (s *API) setCredentials(credentials *api.Credentials) {
	s.credentials = credentials
}

// CreateAPIServiceInstance ...
func CreateAPIServiceInstance(
	credentials *api.Credentials,
	httpClient httpclient.Interface,
	logger logger.Interface,
) api.Interface {
	instance := &API{
		httpclient: httpClient,
		logger:     logger,
		url:        "https://www.ibroker.es/tradingview/api/",
	}
	instance.setCredentials(credentials)
	instance.SetTimeout(time.Second * 10)

	return instance
}
