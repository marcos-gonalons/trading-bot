package modifyposition

import (
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
	"errors"
	"io"
	"net/http"
)

// RequestParameters ...
type RequestParameters struct {
	AccessToken string
	TakeProfit  *string
	StopLoss    *string
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

	err = optionsRequest(endpoint, http.MethodPut)
	if err != nil {
		return
	}

	rq, err := http.NewRequest(
		http.MethodPut,
		endpoint,
		getRequestBody(params.TakeProfit, params.StopLoss),
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	rq.Header.Set("Authorization", "Bearer "+params.AccessToken)
	response, err := httpClient.Do(rq, logger.ModifyPositionRequest)
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

func getRequestBody(takeProfit *string, stopLoss *string) io.Reader {
	body := ""

	if takeProfit != nil && stopLoss != nil {
		body = "" +
			"stopLoss=" + *stopLoss + "&" +
			"takeProfit=" + *takeProfit
	} else {
		if takeProfit != nil {
			body = "takeProfit=" + *takeProfit
		}
		if stopLoss != nil {
			body = "stopLoss=" + *stopLoss
		}
	}

	return utils.GetBodyForHTTPRequest(body)
}
