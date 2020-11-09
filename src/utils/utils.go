package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"
)

// FloatToString ...
func FloatToString(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// IntToString ...
func IntToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

// GetBodyAsString ...
func GetBodyAsString(body io.ReadCloser) string {
	return ""
	if body == nil {
		return ""
	}

	rawBody, _ := ioutil.ReadAll(body)
	body = ioutil.NopCloser(bytes.NewBuffer(rawBody))

	return string(rawBody)
}
