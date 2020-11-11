package utils

import (
	"fmt"
	"strconv"
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
