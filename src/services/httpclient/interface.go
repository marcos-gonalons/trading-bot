package httpclient

import (
	"TradingBot/src/services/logger/types"
	"io"
	"net/http"
	"time"
)

type Interface interface {
	Do(rq *http.Request, logType types.LogType) (*http.Response, error)
	SetTimeout(timeout time.Duration)
	MapJSONResponseToStruct(targetStruct interface{}, responseBody io.Reader) (string, error)
	GetBodyForHTTPRequest(body string) io.Reader
}
