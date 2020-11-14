package utils

import (
	"TradingBot/src/services/api"
	"encoding/json"
	"fmt"
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

// GetWorkingOrders ...
func GetWorkingOrders(orders []*api.Order) []*api.Order {
	var workingOrders []*api.Order
	for _, order := range orders {
		if order.Status == "working" {
			workingOrders = append(workingOrders, order)
		}
	}
	return workingOrders
}
