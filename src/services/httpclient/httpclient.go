package httpclient

import (
	"TradingBot/src/services/logger"
	"TradingBot/src/services/logger/types"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Service wrapper of the original net/http. So it can be used with the interface and easily mocked for unit testing.
type Service struct {
	http.Client
	Logger logger.Interface
}

func (s *Service) Do(rq *http.Request, logType types.LogType) (*http.Response, error) {
	s.Logger.Log("REQUEST -> "+s.getRQStringRepresentation(rq), logType)
	return s.Client.Do(rq)
}

func (s *Service) SetTimeout(timeout time.Duration) {
	s.Client.Timeout = timeout
}

func (s *Service) MapJSONResponseToStruct(targetStruct interface{}, responseBody io.Reader) (string, error) {
	rawBody, _ := ioutil.ReadAll(responseBody)
	responseBody = ioutil.NopCloser(bytes.NewBuffer(rawBody))

	err := json.NewDecoder(responseBody).Decode(&targetStruct)

	return string(rawBody), err
}

func (s *Service) GetBodyForHTTPRequest(body string) io.Reader {
	return bytes.NewBuffer([]byte(body))
}

func (s *Service) getRQStringRepresentation(rq *http.Request) string {
	var bodyAsStr string
	if rq.Body != nil {
		contents, _ := ioutil.ReadAll(rq.Body)
		bodyAsStr = string(contents)
		rq.Body = ioutil.NopCloser(bytes.NewReader(contents))
	} else {
		bodyAsStr = ""
	}

	idk := struct {
		Method  string
		URL     *url.URL
		Body    string
		Headers http.Header
	}{
		Method:  rq.Method,
		URL:     rq.URL,
		Body:    bodyAsStr,
		Headers: rq.Header,
	}

	str, _ := json.Marshal(idk)
	return string(str)
}

func GetInstance(logger logger.Interface) Interface {
	return &Service{
		Logger: logger,
	}
}
