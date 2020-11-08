package createorder

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"time"
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

	url = url + "?requestId=" + getRandomRequestID(10)
	rq, err := http.NewRequest(
		http.MethodPost,
		url,
		getRequestBody(order),
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	rq.Header.Set("Authorization", "Bearer "+accessToken)
	response, err := httpClient.Do(rq, logger.CreateOrderRequest)
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

func getRandomRequestID(length uint) string {
	var src = rand.NewSource(time.Now().UnixNano())
	var letterIdxBits uint = 6
	var letterIdxMask int64 = 1<<letterIdxBits - 1
	var letterIdxMax = 63 / letterIdxBits
	var characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	requestID := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(characters) {
			requestID[i] = characters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(requestID)
}

func getRequestBody(order *api.Order) io.Reader {
	body := "" +
		"currentAsk=" + utils.FloatToString(float64(*order.CurrentAsk)) + "&" +
		"currentBid=" + utils.FloatToString(float64(*order.CurrentBid)) + "&" +
		"durationType=DAY&" +
		"instrument=" + order.Instrument + "&" +
		"side=" + order.Side + "&" +
		"type=" + order.Type + "&" +
		"qty=" + utils.FloatToString(float64(order.Qty))

	if order.StopPrice != nil {
		body = body + "&" + "stopPrice=" + utils.FloatToString(float64(*order.StopPrice))
	}

	if order.LimitPrice != nil {
		body = body + "&" + "limitPrice=" + utils.FloatToString(float64(*order.LimitPrice))
	}

	if order.StopLoss != nil {
		body = body + "&" + "stopLoss=" + utils.FloatToString(float64(*order.StopLoss))
	}

	if order.TakeProfit != nil {
		body = body + "&" + "takeProfit=" + utils.FloatToString(float64(*order.TakeProfit))
	}

	return bytes.NewBuffer([]byte(body))
}
