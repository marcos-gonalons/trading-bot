package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// Service wrapper of the original net/http. So it can be used with the interface and easily mocked for unit testing.
type Service struct {
	http.Client
}

// Do ...
func (s *Service) Do(rq *http.Request) (*http.Response, error) {
	return s.Client.Do(rq)
}

// SetTimeout sets the original net/http client timeout
func (s *Service) SetTimeout(timeout time.Duration) {
	s.Client.Timeout = timeout
}

// MapJSONResponseToStruct does what it says
func (s *Service) MapJSONResponseToStruct(targetStruct interface{}, responseBody io.Reader) ([]byte, error) {
	rawBody, _ := ioutil.ReadAll(responseBody)
	responseBody = ioutil.NopCloser(bytes.NewBuffer(rawBody))

	err := json.NewDecoder(responseBody).Decode(&targetStruct)

	return rawBody, err
}
