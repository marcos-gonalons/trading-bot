package modifyposition

import (
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"bytes"
	"errors"
	"io"
	"net/http"
)

// Request ...
func Request(
	url string,
	httpClient httpclient.Interface,
	accessToken string,
	takeProfit *string,
	stopLoss *string,
	setHeaders func(rq *http.Request),
	optionsRequest func() error,
) (err error) {
	var mappedResponse = &APIResponse{}

	err = optionsRequest()
	if err != nil {
		return
	}

	rq, err := http.NewRequest(
		http.MethodPut,
		url,
		getRequestBody(takeProfit, stopLoss),
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	rq.Header.Set("Authorization", "Bearer "+accessToken)
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

	return bytes.NewBuffer([]byte(body))
}