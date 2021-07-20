package GER30

import "TradingBot/src/types"

var ResistanceBreakoutParams = types.StrategyParams{
	RiskPercentage:                  1,
	StopLossDistance:                24,
	TakeProfitDistance:              34,
	CandlesAmountForHorizontalLevel: 24,
	TPDistanceShortForTighterSL:     0,
	SLDistanceWhenTPIsVeryClose:     0,
	PriceOffset:                     1,
	TrendCandles:                    60,
	TrendDiff:                       15,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{"January", "February", "March", "April", "May", "June", "July", "August", "September"},
		ValidWeekdays:  []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"},
		ValidHalfHours: []string{"9:00", "9:30", "10:00", "10:30", "11:00", "11:30", "12:00", "12:30", "13:00", "13:30", "14:00", "16:00", "16:30", "17:00", "17:30", "20:00", "20:30"},
	},
	MaxTradeExecutionPriceDifference: 3,
}

var SupportBreakoutParams = types.StrategyParams{
	RiskPercentage:                  1,
	StopLossDistance:                15,
	TakeProfitDistance:              34,
	CandlesAmountForHorizontalLevel: 14,
	TPDistanceShortForTighterSL:     1,
	SLDistanceWhenTPIsVeryClose:     0,
	PriceOffset:                     2,
	TrendCandles:                    90,
	TrendDiff:                       30,
	ValidTradingTimes: types.TradingTimes{
		ValidMonths:    []string{"January", "March", "April", "May", "June", "August", "September", "October", "December"},
		ValidWeekdays:  []string{"Monday", "Tuesday", "Thursday", "Friday"},
		ValidHalfHours: []string{"8:00", "8:30", "9:00", "10:00", "10:30", "11:00", "11:30", "12:00", "12:30", "13:00", "14:00", "14:30", "15:00", "15:30", "16:00", "16:30", "17:00", "18:00"},
	},
	MaxTradeExecutionPriceDifference: 3,
}
