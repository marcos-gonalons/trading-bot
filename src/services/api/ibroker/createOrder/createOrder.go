package createorder

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

	err = optionsRequest(endpoint, http.MethodPost)
	if err != nil {
		return
	}

	endpoint = endpoint + "?requestId=" + utils.GetRandomString(10)
	rq, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		getRequestBody(params.Order),
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	rq.Header.Set("Authorization", "Bearer "+params.AccessToken)
	response, err := httpClient.Do(rq, logger.CreateOrderRequest)
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

func getRequestBody(order *api.Order) io.Reader {
	body := "" +
		"currentAsk=" + *order.StringValues.CurrentAsk + "&" +
		"currentBid=" + *order.StringValues.CurrentBid + "&" +
		"durationType=DAY&" +
		"instrument=" + order.Instrument + "&" +
		"side=" + order.Side + "&" +
		"type=" + order.Type + "&" +
		"qty=" + *order.StringValues.Qty

	if order.StopPrice != nil {
		body = body + "&" + "stopPrice=" + *order.StringValues.StopPrice
	}

	if order.LimitPrice != nil {
		body = body + "&" + "limitPrice=" + *order.StringValues.LimitPrice
	}

	if order.StopLoss != nil {
		body = body + "&" + "stopLoss=" + *order.StringValues.StopLoss
	}

	if order.TakeProfit != nil {
		body = body + "&" + "takeProfit=" + *order.StringValues.TakeProfit
	}

	return utils.GetBodyForHTTPRequest(body)
}
