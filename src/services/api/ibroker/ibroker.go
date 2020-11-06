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
	accountID   string
	accessToken *api.AccessToken
}

// SetAccountID ...
func (s *API) SetAccountID(accountID string) {
	s.accountID = accountID
}

// Login ...
func (s *API) Login(username, password string) (accessToken *api.AccessToken, err error) {
	url := s.getURL("authorize")

	rq, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer([]byte("locale=en&login="+username+"&password="+password)),
	)
	if err != nil {
		return
	}

	s.setHeaders(rq)
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
	return "https://www.ibroker.es/tradingview/api/" + endpoint
}

func (s *API) setHeaders(rq *http.Request) {
	rq.Header.Set("Host", "www.ibroker.es")
	rq.Header.Set("Connection", "keep-alive")
	rq.Header.Set("Pragma", "no-cache")
	rq.Header.Set("Cache-Control", "no-cache")
	rq.Header.Set("Accept", "application/json")
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

// CreateAPIServiceInstance ...
func CreateAPIServiceInstance() api.Interface {
	httpclient := &httpclient.Service{}
	httpclient.SetTimeout(time.Second * 5)

	return &API{
		httpclient: httpclient,
	}
}
