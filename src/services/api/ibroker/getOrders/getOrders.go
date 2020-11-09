package getorders

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"errors"
	"net/http"
)

// Request ...
func Request(
	url string,
	httpClient httpclient.Interface,
	accessToken string,
	setHeaders func(rq *http.Request),
	optionsRequest func() error,
) (orders []*api.Order, err error) {
	var mappedResponse = &APIResponse{}

	err = optionsRequest()
	if err != nil {
		return
	}

	rq, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	rq.Header.Set("Authorization", "Bearer "+accessToken)
	response, err := httpClient.Do(rq, logger.GetOrdersRequest)
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

	for _, responseOrder := range mappedResponse.Data {
		limitPrice := responseOrder.LimitPrice
		stopPrice := responseOrder.StopPrice
		parentID := responseOrder.ParentID
		order := &api.Order{
			ID:         responseOrder.ID,
			Instrument: responseOrder.Instrument,
			Qty:        responseOrder.Qty,
			Side:       responseOrder.Side,
			Type:       responseOrder.Type,
			Status:     responseOrder.Status,
			ParentID:   parentID,
			LimitPrice: &limitPrice,
			StopPrice:  &stopPrice,
		}
		orders = append(orders, order)
	}

	return
}
