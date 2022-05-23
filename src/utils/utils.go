package utils

import (
	"TradingBot/src/constants"
	"TradingBot/src/services/api"
	"TradingBot/src/types"
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

// FloatToString ...
func FloatToString(v float64, decimals int64) string {
	m := math.Pow(10, float64(decimals))
	v = math.Round(v*m) / m
	decimalsAsString := IntToString(decimals)
	return fmt.Sprintf("%."+decimalsAsString+"f", v)
}

// StringToFloat ...
func StringToFloat(v string) float64 {
	n, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return n
	}
	return 0.0
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

// GetTimestamp ...
func GetTimestamp(t time.Time, timeLayout string) int64 {
	dateString := t.Format("2006-01-02 " + timeLayout)
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

// GetCurrentTimeHourAndMinutes ...
func GetCurrentTimeHourAndMinutes() (int, int) {
	n := time.Now()
	t := n.Add(time.Minute * -1)

	currentHour, _ := strconv.Atoi(t.Format("15"))
	currentMinutes, _ := strconv.Atoi(t.Format("04"))

	return currentHour, currentMinutes
}

// IsNowWithinTradingHours ...
func IsNowWithinTradingHours(symbol *types.Symbol) bool {
	if symbol.MarketType == constants.ForexType {
		if IsOutsideForexHours(time.Now()) {
			return false
		}
	} else {
		if IsNowWeekend() && !symbol.TradeableOnWeekends {
			return false
		}
	}

	if symbol.TradingHours.Start == symbol.TradingHours.End {
		return true
	}

	var currentHourTime, timeStartingHour, timeEndingHour time.Time

	currentHour, _ := GetCurrentTimeHourAndMinutes()
	currentHourTime, _ = time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:00:00"))

	timeStartingHour = time.Date(currentHourTime.Year(), currentHourTime.Month(), currentHourTime.Day(), int(symbol.TradingHours.Start), 0, 0, 0, currentHourTime.Location())
	timeEndingHour = time.Date(currentHourTime.Year(), currentHourTime.Month(), currentHourTime.Day(), int(symbol.TradingHours.End), 0, 0, 0, currentHourTime.Location())

	if symbol.TradingHours.End < symbol.TradingHours.Start {
		if currentHour < int(symbol.TradingHours.End) {
			timeStartingHour = timeStartingHour.Add(-24 * time.Hour)
		} else if currentHour >= int(symbol.TradingHours.Start) {
			timeEndingHour = timeEndingHour.Add(24 * time.Hour)
		}
	}

	timeEndingHour = timeEndingHour.Add(-3 * time.Minute)
	currentHourTime = time.Date(currentHourTime.Year(), currentHourTime.Month(), currentHourTime.Day(), currentHourTime.Hour(), time.Now().Minute(), 0, 0, currentHourTime.Location())

	return currentHourTime.Unix() >= timeStartingHour.Unix() && currentHourTime.Unix() < timeEndingHour.Unix()
}

// IsNowWeekend ...
func IsNowWeekend() bool {
	weekDay := time.Now().Weekday()
	return weekDay == 0 || weekDay == 6
}

// FilterOrdersBySymbol ...
func FilterOrdersBySymbol(orders []*api.Order, symbol string) []*api.Order {
	filteredOrders := funk.Filter(orders, func(o *api.Order) bool {
		return o.Instrument == symbol
	})

	return filteredOrders.([]*api.Order)
}

// FilterOrdersBySymbol ...
func FilterOrdersBySide(orders []*api.Order, side string) []*api.Order {
	filteredOrders := funk.Filter(orders, func(o *api.Order) bool {
		return o.Side == side
	})

	return filteredOrders.([]*api.Order)
}

// FindPositionBySymbol ...
func FindPositionBySymbol(positions []*api.Position, symbol string) *api.Position {
	p := funk.Find(positions, func(p *api.Position) bool {
		return p.Instrument == symbol
	})

	if p == nil {
		return nil
	}

	return p.(*api.Position)
}

// IsExecutionTimeValid ...
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

func IsOutsideForexHours(t time.Time) bool {
	// t is a timestamp in spanish time
	weekday := t.Weekday()

	// sun 0, mon 1, tue 2, wed 3, thu 4, fri 5, sat 6
	if weekday >= 1 && weekday <= 4 {
		return false
	}

	if weekday == 6 { // saturday
		return true
	}

	hour, _ := strconv.Atoi(t.Format("15"))
	if weekday == 5 { // friday
		return hour > 22
	}

	if weekday == 0 { // sunday
		return hour < 23
	}

	return false
}

func GetPositionSize(
	currentBalance float64,
	riskPercentage float64,
	stopLossDistance float64,
	minPositionSize float64,
	eurExchangeRate float64,
) float32 {
	size := math.Floor((currentBalance*(riskPercentage/100))/(stopLossDistance*minPositionSize*eurExchangeRate)) * minPositionSize
	if size == 0 {
		size = minPositionSize
	}
	return float32(size)
}
