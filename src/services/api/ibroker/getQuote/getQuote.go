package getquote

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"errors"
	"net/http"
)

// RequestParameters ...
type RequestParameters struct {
	AccessToken string
	AccountID   string
	Symbol      string
}

// Request ...
func Request(
	endpoint string,
	httpClient httpclient.Interface,
	setHeaders func(rq *http.Request),
	optionsRequest func(url string, httpMethod string) error,
	params *RequestParameters,
) (quote *api.Quote, err error) {
	var mappedResponse = &APIResponse{}

	err = optionsRequest(endpoint, http.MethodGet)
	if err != nil {
		return
	}

	endpoint = endpoint + "?symbols=" + params.Symbol + "&accountId=" + params.AccountID
	rq, err := http.NewRequest(
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	rq.Header.Set("Authorization", "Bearer "+params.AccessToken)
	response, err := httpClient.Do(rq, logger.GetQuoteRequest)
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

	if len(mappedResponse.Data) == 0 {
		err = errors.New("Empty quotes array - Response was " + rawBody)
		return
	}

	if mappedResponse.Data[0].Status != "ok" {
		err = errors.New("Bad status - Response was " + rawBody)
		return
	}

	emptyValueStruct := QuoteValue{}
	if mappedResponse.Data[0].Value == emptyValueStruct {
		err = errors.New("Response does not contain the quote - Response was " + rawBody)
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
