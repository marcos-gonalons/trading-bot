package ibroker

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
)

// API ...
type API struct {
	httpclient  httpclient.Interface
	accessToken *api.AccessToken
	credentials *api.Credentials
}

// Login ...
func (s *API) Login() (accessToken *api.AccessToken, err error) {
	url := s.getURL("authorize")

	err = s.doOptionsRequest(url, "GET")
	if err != nil {
		return
	}

	rq, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer([]byte("locale=en&login="+s.credentials.Username+"&password="+s.credentials.Password)),
	)
	if err != nil {
		return
	}

	s.setHeaders(rq, false, "")
	response, err := s.httpclient.Do(rq)
	if err != nil {
		return
	}

	mappedResponse := &LoginResponse{}
	_, err = s.httpclient.MapJSONResponseToStruct(mappedResponse, response.Body)
	if err != nil {
		return
	}

	if mappedResponse.ErrorMsg != "" {
		err = errors.New("Api error -> " + mappedResponse.ErrorMsg)
		return
	}

	if mappedResponse.Data.AccessToken == "" {
		err = errors.New("Empty access token")
		return
	}

	accessToken = &api.AccessToken{
		Token:      mappedResponse.Data.AccessToken,
		Expiration: time.Unix(int64(mappedResponse.Data.Expiration), 0),
	}

	s.accessToken = accessToken
	return
}

func (s *API) getURL(endpoint string) string {
	// todo: env var for the url?
	return "https://marcos.free.beeceptor.com/" + endpoint
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

func (s *API) doOptionsRequest(url string, method string) (err error) {
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
	_, err = s.httpclient.Do(rq)
	if err != nil {
		return
	}

	return
}

func (s *API) setCredentials(credentials *api.Credentials) {
	s.credentials = credentials
}

// CreateAPIServiceInstance ...
func CreateAPIServiceInstance(credentials *api.Credentials) api.Interface {
	httpclient := &httpclient.Service{}
	httpclient.SetTimeout(time.Second * 5)

	instance := &API{
		httpclient: httpclient,
	}
	instance.setCredentials(credentials)

	return instance
}
