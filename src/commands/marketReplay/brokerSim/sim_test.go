package brokerSim

import (
	"TradingBot/src/types"
	"testing"
)

func TestIsPriceWithinCandle(t *testing.T) {
	testCase1 := make(map[string]interface{})
	testCase1["price"] = 100.
	testCase1["candleLow"] = 90.
	testCase1["candleHigh"] = 110.
	testCase1["expectedResult"] = true

	testCase2 := make(map[string]interface{})
	testCase2["price"] = 100.
	testCase2["candleLow"] = 90.
	testCase2["candleHigh"] = 95.
	testCase2["expectedResult"] = false

	testCase3 := make(map[string]interface{})
	testCase3["price"] = 100.
	testCase3["candleLow"] = 100.
	testCase3["candleHigh"] = 100.
	testCase3["expectedResult"] = true

	testCases := []map[string]interface{}{
		testCase1,
		testCase2,
		testCase3,
	}

	for _, testCase := range testCases {
		candle := &types.Candle{
			Low:  testCase["candleLow"].(float64),
			High: testCase["candleHigh"].(float64),
		}

		result := isPriceWithinCandle(testCase["price"].(float64), candle)
		if result != testCase["expectedResult"] {
			t.Error(testCase)
		}
	}

}

func TestgetOrderExecutionPrice(t *testing.T) {

}
