package login

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"encoding/json"

	"bytes"
	"errors"
	"net/http"
	"time"
)

// Request ...
func Request(
	url string,
	credentials *api.Credentials,
	httpClient httpclient.Interface,
	setHeaders func(rq *http.Request),
	optionsRequest func() error,
) (accessToken *api.AccessToken, err error) {
	var mappedResponse = &APIResponse{}

	err = optionsRequest()
	if err != nil {
		return
	}

	rq, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer([]byte("locale=en&login="+credentials.Username+"&password="+credentials.Password)),
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	response, err := httpClient.Do(rq, logger.LoginRequest)
	if err != nil {
		return
	}

	_, err = httpClient.MapJSONResponseToStruct(mappedResponse, response.Body)
	if err != nil {
		return
	}

	str, err := json.Marshal(mappedResponse)
	if err != nil {
		return
	}
	responseAsString := string(str)

	if mappedResponse.ErrorMsg != "" {
		err = errors.New("Api error -> " + mappedResponse.ErrorMsg)
		return
	}

	if mappedResponse.Data.AccessToken == "" {
		err = errors.New("Empty access token - Response was " + responseAsString)
		return
	}

	accessToken = &api.AccessToken{
		Token:      mappedResponse.Data.AccessToken,
		Expiration: time.Unix(int64(mappedResponse.Data.Expiration), 0),
	}

	return
}
