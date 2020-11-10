package mainstrategy

import (
	"encoding/json"
	"time"
)

func (s *Strategy) updateCandles(currentExecutionTime time.Time) {
	currentMinutes := currentExecutionTime.Format("04")
	previousMinutes := s.previousExecutionTime.Format("04")

	currentPrice := s.quote.Price

	if currentMinutes != previousMinutes {
		lastCandle, _ := json.Marshal(s.candles[len(s.candles)-1])
		s.Logger.Log("Adding new candle to the candles array - Last candle was " + string(lastCandle))
		s.candles = append(s.candles, &Candle{
			Open:      currentPrice,
			Low:       currentPrice,
			High:      currentPrice,
			Volume:    s.quote.Volume,
			Timestamp: getTimestampWith0Seconds(currentExecutionTime),
		})
	} else {
		index := len(s.candles) - 1
		if currentPrice <= s.candles[index].Low {
			s.candles[index].Low = currentPrice
		}
		if currentPrice >= s.candles[index].High {
			s.candles[index].High = currentPrice
		}
		s.candles[index].Close = currentPrice
		s.candles[index].Volume = s.quote.Volume
		if s.candles[index].Timestamp == 0 {
			s.candles[index].Timestamp = getTimestampWith0Seconds(currentExecutionTime)
		}
	}

	/***
		Remember to add logs

		Very, very important
		Round ALL the prices used in ALL the calls to 2 decimals. Otherwise it won't work.

		When creating an order, I need to save the 3 created orders somewhere (the limit/stop order, it's sl and it's tp)
		The SL and the TP will have the parentID of the main one. The main one will have the parentID null
		All 3 orders will have the status "working".

		When modifying an order that hasn't been filled yet, I can use the ID of the main order to change it's sl, tp, or it's limit/stop price.
		When modifying the sl/tp of a position, I need to use the ID of the sl/tp order.
		Or I can just use the modifyposition api


		Take into consideration
		Let's say the bot dies, for whatever reason, at 15:00pm
		I revive him at 15:05
		It will have lost all the candles[]

		To mitigate this
		As I add a candle to the candles[]
		Save the candles to the csv file
		When booting the bot; initialize the candles array with those in the csv file


		When booting {
			if !csv file, create the csv file
			else, load candles[] from the file
		}
	***/

}

func getTimestampWith0Seconds(t time.Time) int64 {
	dateString := t.Format("2006-01-02 15:04:00")
	date, _ := time.Parse("2006-01-02 15:04:05", dateString)
	return date.Unix()
}
