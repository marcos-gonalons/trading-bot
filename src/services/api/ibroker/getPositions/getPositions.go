package getpositions

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"TradingBot/src/utils"
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

	responseAsString := utils.GetBodyAsString(response.Body)
	_, err = httpClient.MapJSONResponseToStruct(mappedResponse, response.Body)
	if err != nil {
		errorMessage := "" +
			"Error while mapping JSON - " + err.Error() +
			"\n Response was - " + responseAsString
		err = errors.New(errorMessage)
		return
	}

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
