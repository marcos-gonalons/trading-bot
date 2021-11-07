package modifyorder

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
	logger "TradingBot/src/services/logger/types"
	"TradingBot/src/utils"
	"errors"
	"io"
	"net/http"
)

// RequestParameters ...
type RequestParameters struct {
	AccessToken  string
	Order        *api.Order
	IsLimitOrder func(order *api.Order) bool
	IsStopOrder  func(order *api.Order) bool
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
		getRequestBody(params),
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
		err = errors.New("Api error -> " + mappedResponse.ErrorMsg + "\n Raw Body is -> " + rawBody)
		return
	}

	if mappedResponse.Status != "ok" {
		err = errors.New("Bad status - Response was " + rawBody)
		return
	}

	return
}

func getRequestBody(params *RequestParameters) io.Reader {
	order := params.Order

	body := "" +
		"currentAsk=" + *order.StringValues.CurrentAsk + "&" +
		"currentBid=" + *order.StringValues.CurrentBid + "&" +
		"durationType=DAY&" +
		"qty=" + *order.StringValues.Qty + "&" +
		"id=" + order.ID

	if params.IsLimitOrder(order) {
		body = body + "&" + "limitPrice=" + *order.StringValues.LimitPrice
		body = body + "&" + "stopPrice=0"
	}

	if params.IsStopOrder(order) {
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
