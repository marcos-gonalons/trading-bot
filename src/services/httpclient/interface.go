package httpclient

import (
	"io"
	"net/http"
	"time"
)

// Interface ...
type Interface interface {
	Do(rq *http.Request) (*http.Response, error)
	SetTimeout(timeout time.Duration)
	MapJSONResponseToStruct(targetStruct interface{}, responseBody io.Reader) ([]byte, error)
}
