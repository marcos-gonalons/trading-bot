package closeposition

import (
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"errors"
	"net/http"
)

// RequestParameters ...
type RequestParameters struct {
	AccessToken string
}

// Request ...
func Request(
	endpoint string,
	httpClient httpclient.Interface,
	setHeaders func(rq *http.Request),
	optionsRequest func(url string, httpMethod string) error,
	params *RequestParameters,
) (n interface{}, err error) {
	var mappedResponse = &APIResponse{}

	err = optionsRequest(endpoint, http.MethodDelete)
	if err != nil {
		return
	}

	rq, err := http.NewRequest(
		http.MethodDelete,
		endpoint,
		nil,
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	rq.Header.Set("Authorization", "Bearer "+params.AccessToken)
	response, err := httpClient.Do(rq, logger.ClosePositionRequest)
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
		err = errors.New("Api error -> " + mappedResponse.ErrorMsg)
		return
	}

	if mappedResponse.Status != "ok" {
		err = errors.New("Bad status - Response was " + rawBody)
		return
	}

	return
}
