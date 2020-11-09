package getquote

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
	accountID string,
	symbol string,
	setHeaders func(rq *http.Request),
	optionsRequest func() error,
) (quote *api.Quote, err error) {
	var mappedResponse = &APIResponse{}

	err = optionsRequest()
	if err != nil {
		return
	}

	url = url + "?symbols=" + symbol + "&accountId=" + accountID
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
	response, err := httpClient.Do(rq, logger.GetQuoteRequest)
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

	if len(mappedResponse.Data) == 0 {
		err = errors.New("Empty quotes array - Response was " + responseAsString)
		return
	}

	if mappedResponse.Data[0].Status != "ok" {
		err = errors.New("Bad status - Response was " + responseAsString)
		return
	}

	emptyValueStruct := QuoteValue{}
	if mappedResponse.Data[0].Value == emptyValueStruct {
		err = errors.New("Response does not contain the quote - Response was " + responseAsString)
		return
	}

	quote = &api.Quote{
		Ask:    mappedResponse.Data[0].Value.Ask,
		Bid:    mappedResponse.Data[0].Value.Bid,
		Price:  mappedResponse.Data[0].Value.CurrentPrice,
		Volume: mappedResponse.Data[0].Value.Volume,
	}
	return
}
