package getpositions

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"encoding/json"
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
) (positions []*api.Position, err error) {
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
	response, err := httpClient.Do(rq, logger.GetPositionsRequest)
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

	for _, responsePosition := range mappedResponse.Data {
		position := &api.Position{
			Instrument:   responsePosition.Instrument,
			Qty:          responsePosition.Qty,
			Side:         responsePosition.Side,
			AvgPrice:     responsePosition.AvgPrice,
			UnrealizedPl: responsePosition.UnrealizedPL,
		}
		positions = append(positions, position)
	}

	return
}
