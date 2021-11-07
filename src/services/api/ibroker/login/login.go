package login

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
	logger "TradingBot/src/services/logger/types"
	"TradingBot/src/utils"

	"errors"
	"net/http"
	"time"
)

// RequestParameters ...
type RequestParameters struct {
	Credentials *api.Credentials
}

// Request ...
func Request(
	endpoint string,
	httpClient httpclient.Interface,
	setHeaders func(rq *http.Request),
	optionsRequest func(url string, httpMethod string) error,
	params *RequestParameters,
) (accessToken *api.AccessToken, err error) {
	var mappedResponse = &APIResponse{}

	err = optionsRequest(endpoint, http.MethodPost)
	if err != nil {
		return
	}

	bodyString := "locale=en&login=" + params.Credentials.Username + "&password=" + params.Credentials.Password
	rq, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		utils.GetBodyForHTTPRequest(bodyString),
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	response, err := httpClient.Do(rq, logger.LoginRequest)
	if err != nil {
		return
	}

	rawBody, err := httpClient.MapJSONResponseToStruct(mappedResponse, response.Body)
	if err != nil {
		errorMessage := "" +
			"Error while mapping JSON - " + err.Error() +
			"\n Response was - " + rawBody
		err = errors.New(errorMessage)
		return
	}

	if mappedResponse.ErrorMsg != "" {
		err = errors.New("Api error -> " + mappedResponse.ErrorMsg + "\n Raw Body is -> " + rawBody)
		return
	}

	if mappedResponse.Data.AccessToken == "" {
		err = errors.New("Empty access token - Response was " + rawBody)
		return
	}

	accessToken = &api.AccessToken{
		Token:      mappedResponse.Data.AccessToken,
		Expiration: time.Unix(int64(mappedResponse.Data.Expiration), 0),
	}

	return
}
