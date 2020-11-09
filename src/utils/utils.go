package utils

import (
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
