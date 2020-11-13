package mainstrategy

import (
	"TradingBot/src/utils"
	"encoding/json"
	"os"
	"strconv"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper"
)

func (s *Strategy) updateCandles(currentExecutionTime time.Time, data *tradingviewsocket.QuoteData) {
	if data.Price == nil && data.Volume == nil {
		return
	}

	currentMinutes := currentExecutionTime.Format("04")
	previousMinutes := s.previousExecutionTime.Format("04")

	var currentPrice float64
	if data.Price != nil {
		currentPrice = *data.Price
	} else {
		currentPrice = 0
	}

	if currentMinutes != previousMinutes {
		s.updateCSVWithLastCandle()
		lastCandle, _ := json.Marshal(s.candles[len(s.candles)-1])
		s.Logger.Log("Adding new candle to the candles array -> " + string(lastCandle))
		s.candles = append(s.candles, &Candle{
			Open:      currentPrice,
			Low:       currentPrice,
			High:      currentPrice,
			Volume:    *data.Volume - s.currentVolume,
			Timestamp: getTimestampWith0Seconds(currentExecutionTime),
		})
	} else {
		index := len(s.candles) - 1
		if data.Price != nil {
			if currentPrice <= s.candles[index].Low {
				s.candles[index].Low = currentPrice
			}
			if currentPrice >= s.candles[index].High {
				s.candles[index].High = currentPrice
			}
			s.candles[index].Close = currentPrice
		}
		s.candles[index].Volume += *data.Volume - s.currentVolume
		if s.candles[index].Timestamp == 0 {
			s.candles[index].Timestamp = getTimestampWith0Seconds(currentExecutionTime)
		}
	}

	/***
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

		Very important as well
		When closing a position, and the position has TP and SL; IBROKER WILL NOT LET YOU
		CLOSE THE POSITION UNTIL YOU CLOSE THE TP AND SL FIRST
		So; if the script needs to close a position, CLOSE THE SL AND TP FIRST.
	***/

}

func (s *Strategy) initCandles() {
	s.candles = nil
	s.candles = []*Candle{&Candle{}}

	/**
		todo: use go routine to save the candles csv file??
	**/
	now := time.Now()
	s.csvFileName = now.Format("2006-01-02") + "-candles.csv"
	csvFile, err := os.OpenFile(s.csvFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	defer csvFile.Close()
	if err != nil {
		s.csvFileMtx.Lock()
		csvFile, err = os.Create(s.csvFileName)
		defer csvFile.Close()
		if err != nil {
			s.csvFileMtx.Unlock()
			panic("Error while creating the csv file")
		} else {
			csvFile.Write([]byte("Time,Open,High,Low,Close,Volume\n"))
			s.csvFileMtx.Unlock()
		}
	}

}

func (s *Strategy) updateCSVWithLastCandle() {
	lastCandle := s.candles[len(s.candles)-1]
	if lastCandle.Timestamp == 0 {
		return
	}
	row := "" +
		strconv.FormatInt(lastCandle.Timestamp, 10) + "," +
		utils.FloatToString(float64(lastCandle.Open), 5) + "," +
		utils.FloatToString(float64(lastCandle.High), 5) + "," +
		utils.FloatToString(float64(lastCandle.Low), 5) + "," +
		utils.FloatToString(float64(lastCandle.Close), 5) + "," +
		utils.FloatToString(lastCandle.Volume, 5) + "\n"

	var err error

	csvFile, err := os.OpenFile(s.csvFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	defer csvFile.Close()

	s.csvFileMtx.Lock()
	_, err = csvFile.Write([]byte(row))
	if err != nil {
		s.Logger.Log("Error when writting the last candle in the csv file -> " + err.Error())
	}
	s.csvFileMtx.Unlock()
}

func getTimestampWith0Seconds(t time.Time) int64 {
	dateString := t.Format("2006-01-02 15:04:00")
	date, _ := time.Parse("2006-01-02 15:04:05", dateString)
	return date.Unix()
}
