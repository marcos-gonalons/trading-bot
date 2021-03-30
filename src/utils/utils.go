package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"time"
)

// FloatToString ...
func FloatToString(v float64, decimals int64) string {
	decimalsAsString := IntToString(decimals)
	return fmt.Sprintf("%."+decimalsAsString+"f", v)
}

// IntToString ...
func IntToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

// GetRandomString ...
func GetRandomString(length int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	var characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var letterIdxBits int = 6
	var letterIdxMask int64 = 1<<letterIdxBits - 1
	var letterIdxMax = 63 / letterIdxBits

	requestID := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(characters) {
			requestID[i] = characters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(requestID)
}

// GetStringRepresentation ...
func GetStringRepresentation(data interface{}) string {
	str, _ := json.Marshal(data)
	return string(str)
}

// RepeatUntilSuccess ...
func RepeatUntilSuccess(
	processName string,
	process func() error,
	delayBetweenRetries time.Duration,
	maxRetries uint,
	successCallback func(),
) {
	var retries uint
	retries = 0
	for {
		err := process()

		if maxRetries == 0 {
			if err == nil && successCallback != nil {
				successCallback()
			}
			return
		}

		if err == nil {
			break
		}

		retries++
		if retries == maxRetries {
			panic("There is something wrong while doing " + processName)
		}
		time.Sleep(delayBetweenRetries)
	}
	if successCallback != nil {
		successCallback()
	}
}

// GetTimestampWith0Seconds ...
func GetTimestampWith0Seconds(t time.Time) int64 {
	dateString := t.Format("2006-01-02 15:04:00")
	date, _ := time.Parse("2006-01-02 15:04:05", dateString)
	return date.Unix()
}

// IsInArray ...
func IsInArray(element string, arr []string) bool {
	for _, el := range arr {
		if element == el {
			return true
		}
	}
	return false
}

// GetBodyForHTTPRequest ...
func GetBodyForHTTPRequest(body string) io.Reader {
	return bytes.NewBuffer([]byte(body))
}
