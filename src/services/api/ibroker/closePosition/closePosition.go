package closeposition

import (
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"net/http"
)

// Request ...
func Request(
	url string,
	httpClient httpclient.Interface,
	accessToken string,
	setHeaders func(rq *http.Request),
	optionsRequest func() error,
) (err error) {
	err = optionsRequest()
	if err != nil {
		return
	}

	rq, err := http.NewRequest(
		http.MethodDelete,
		url,
		nil,
	)
	if err != nil {
		return
	}

	setHeaders(rq)
	rq.Header.Set("Authorization", "Bearer "+accessToken)
	_, err = httpClient.Do(rq, logger.ClosePositionRequest)
	if err != nil {
		return
	}

	// TODO: Check the close position response in chrome's console

	return
}
