package ibroker

import (
	"net/http"
	"net/url"
	"time"

	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
)

// API ...
type API struct {
	httpclient httpclient.Interface
	accountID  string
}

// SetAccountID ...
func (s *API) SetAccountID(accountID string) {
	s.accountID = accountID
}

// Login ...
func (s *API) Login(username, password string) (*api.AccessToken, error) {
	mappedResponse := &LoginResponse{}

	rq := &http.Request{
		Method: "POST",
		URL:    &url.URL{},
	}

	response, err := s.httpclient.Do(rq)

	_, err = s.httpclient.MapJSONResponseToStruct(mappedResponse, response.Body)

	if err != nil {
		return nil, err
	}

	return &api.AccessToken{
		Token:      mappedResponse.Data.AccessToken,
		Expiration: time.Unix(mappedResponse.Data.Expiration, 0),
	}, nil
}

// CreateInstance ...
func CreateInstance() interface{} {
	httpclient := &httpclient.Service{}
	httpclient.SetTimeout(time.Second * 5)

	return &API{
		httpclient: httpclient,
	}
}
