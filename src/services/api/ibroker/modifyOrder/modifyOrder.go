package modifyorder

import (
	"TradingBot/src/services/api"
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
	Order       *api.Order
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
		getRequestBody(params.Order),
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	rq.Header.Set("Authorization", "Bearer "+params.AccessToken)
	response, err := httpClient.Do(rq, logger.ModifyOrderRequest)
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

func getRequestBody(order *api.Order) io.Reader {
	body := "" +
		"currentAsk=" + *order.StringValues.CurrentAsk + "&" +
		"currentBid=" + *order.StringValues.CurrentBid + "&" +
		"durationType=DAY&" +
		"qty=" + *order.StringValues.Qty + "&" +
		"id=" + order.ID

	if order.Type == "limit" {
		body = body + "&" + "limitPrice=" + *order.StringValues.LimitPrice
		body = body + "&" + "stopPrice=0"
	}

	if order.Type == "stop" {
		body = body + "&" + "stopPrice=" + *order.StringValues.StopPrice
		body = body + "&" + "limitPrice=0"
	}

	if order.StopLoss != nil {
		body = body + "&" + "stopLoss=" + *order.StringValues.StopLoss
	}

	if order.TakeProfit != nil {
		body = body + "&" + "takeProfit=" + *order.StringValues.TakeProfit
	}

	return utils.GetBodyForHTTPRequest(body)
}
