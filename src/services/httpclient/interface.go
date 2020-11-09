package httpclient

import (
	"TradingBot/src/services/logger"
	"io"
	"net/http"
	"time"
)

// Interface ...
type Interface interface {
	Do(rq *http.Request, logType logger.LogType) (*http.Response, error)
	SetTimeout(timeout time.Duration)
	MapJSONResponseToStruct(targetStruct interface{}, responseBody io.Reader) (string, error)
}
