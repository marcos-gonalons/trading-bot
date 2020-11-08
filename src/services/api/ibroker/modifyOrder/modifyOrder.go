package modifyorder

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Request ...
func Request(
	url string,
	httpClient httpclient.Interface,
	accessToken string,
	order *api.Order,
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
		getRequestBody(order),
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	rq.Header.Set("Authorization", "Bearer "+accessToken)
	response, err := httpClient.Do(rq, logger.ModifyOrderRequest)
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

	if mappedResponse.Status != "ok" {
		err = errors.New("Bad status - Response was " + responseAsString)
		return
	}

	return
}

func getRequestBody(order *api.Order) io.Reader {
	body := "" +
		"currentAsk=" + utils.FloatToString(float64(*order.CurrentAsk)) + "&" +
		"currentBid=" + utils.FloatToString(float64(*order.CurrentBid)) + "&" +
		"durationType=DAY&" +
		"qty=" + utils.FloatToString(float64(order.Qty)) +
		"id=" + utils.IntToString(order.ID)

	if order.Type == "limit" {
		body = body + "&" + "limitPrice=" + utils.FloatToString(float64(*order.LimitPrice))
		body = body + "&" + "stopPrice=0"
	}

	if order.Type == "stop" {
		body = body + "&" + "stopPrice=" + utils.FloatToString(float64(*order.StopPrice))
		body = body + "&" + "limitPrice=0"
	}

	if order.StopLoss != nil {
		body = body + "&" + "stopLoss=" + utils.FloatToString(float64(*order.StopLoss))
	}

	if order.TakeProfit != nil {
		body = body + "&" + "takeProfit=" + utils.FloatToString(float64(*order.TakeProfit))
	}

	return bytes.NewBuffer([]byte(body))
}
