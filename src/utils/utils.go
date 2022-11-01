package utils

import (
	"TradingBot/src/services/api"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strconv"
	"time"

	funk "github.com/thoas/go-funk"
)

func FloatToString(v float64, decimals int64) string {
	m := math.Pow(10, float64(decimals))
	v = math.Round(v*m) / m
	decimalsAsString := IntToString(decimals)
	return fmt.Sprintf("%."+decimalsAsString+"f", v)
}

func StringToFloat(v string) float64 {
	n, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return n
	}
	return 0.0
}

func IntToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

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

func GetStringRepresentation(data interface{}) string {
	str, _ := json.Marshal(data)
	return string(str)
}

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

func GetTimestamp(t time.Time, timeLayout string) int64 {
	dateString := t.Format("2006-01-02 " + timeLayout)
	date, _ := time.Parse("2006-01-02 15:04:05", dateString)
	return date.Unix()
}

func IsInArray(element string, arr []string) bool {
	for _, el := range arr {
		if element == el {
			return true
		}
	}
	return false
}

func GetBodyForHTTPRequest(body string) io.Reader {
	return bytes.NewBuffer([]byte(body))
}

func GetCurrentTimeHourAndMinutes() (int, int) {
	n := time.Now()
	t := n.Add(time.Minute * -1)

	currentHour, _ := strconv.Atoi(t.Format("15"))
	currentMinutes, _ := strconv.Atoi(t.Format("04"))

	return currentHour, currentMinutes
}

func IsWithinTradingHours(utcTime time.Time, tradingHours map[int][]int) bool {
	weekday := int(utcTime.Weekday())
	hour, _ := strconv.Atoi(utcTime.Format("15"))

	for _, h := range tradingHours[weekday] {
		if h == hour {
			return true
		}
	}

	return false
}

func FilterOrdersByMarket(orders []*api.Order, marketName string) []*api.Order {
	filteredOrders := funk.Filter(orders, func(o *api.Order) bool {
		return o.Instrument == marketName
	})

	return filteredOrders.([]*api.Order)
}

func FilterOrdersBySide(orders []*api.Order, side string) []*api.Order {
	filteredOrders := funk.Filter(orders, func(o *api.Order) bool {
		return o.Side == side
	})

	return filteredOrders.([]*api.Order)
}

func FindPositionByMarket(positions []*api.Position, marketName string) *api.Position {
	p := funk.Find(positions, func(p *api.Position) bool {
		return p.Instrument == marketName
	})

	if p == nil {
		return nil
	}

	return p.(*api.Position)
}

func IsExecutionTimeValid(
	t time.Time,
	validMonths []string,
	validWeekDays []string,
	validHalfHours []string,
) bool {
	if len(validMonths) > 0 {
		if !IsInArray(t.Format("January"), validMonths) {
			return false
		}
	}

	if len(validWeekDays) > 0 {
		if !IsInArray(t.Format("Monday"), validWeekDays) {
			return false
		}
	}

	if len(validHalfHours) > 0 {
		currentHour, currentMinutes := GetCurrentTimeHourAndMinutes()
		if currentMinutes >= 30 {
			currentMinutes = 30
		} else {
			currentMinutes = 0
		}

		currentHourString := strconv.Itoa(currentHour)
		currentMinutesString := strconv.Itoa(currentMinutes)
		if len(currentMinutesString) == 1 {
			currentMinutesString += "0"
		}

		return IsInArray(currentHourString+":"+currentMinutesString, validHalfHours)
	}

	return true
}

func GetForexUTCTradingHours() map[int][]int {
	tradingHoursUTC := make(map[int][]int)

	// Monday
	tradingHoursUTC[1] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}
	// Tuesday
	tradingHoursUTC[2] = tradingHoursUTC[1]
	// Wednesday
	tradingHoursUTC[3] = tradingHoursUTC[1]
	// Thursday
	tradingHoursUTC[4] = tradingHoursUTC[1]
	// Friday
	tradingHoursUTC[5] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	// Saturday
	tradingHoursUTC[6] = []int{}
	// Sunday
	tradingHoursUTC[0] = []int{21, 22, 23}

	return tradingHoursUTC
}
