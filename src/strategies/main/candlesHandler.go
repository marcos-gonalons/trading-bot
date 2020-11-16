package mainstrategy

import (
	"TradingBot/src/utils"
	"encoding/json"
	"os"
	"strconv"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper"
)

func (s *Strategy) updateCandles(data *tradingviewsocket.QuoteData) {
	currentMinutes := s.currentExecutionTime.Format("04")
	previousMinutes := s.previousExecutionTime.Format("04")

	var currentPrice float64
	if data.Price != nil {
		currentPrice = *data.Price
	} else {
		currentPrice = 0
	}

	var volume float64
	if data.Volume != nil {
		volume = *data.Volume - s.lastVolume
		if volume < 0 {
			volume = *data.Volume
		}
	} else {
		volume = 0
	}

	if currentMinutes != previousMinutes {
		s.updateCSVWithLastCandle()
		lastCandle, _ := json.Marshal(s.getLastCandle())
		s.Logger.Log("Adding new candle to the candles array -> " + string(lastCandle))
		s.candles = append(s.candles, &Candle{
			Open:      s.getLastCandle().Close,
			Low:       currentPrice,
			High:      currentPrice,
			Volume:    volume,
			Timestamp: utils.GetTimestampWith0Seconds(s.currentExecutionTime),
		})
	} else {
		index := len(s.candles) - 1
		if data.Price != nil {
			if s.candles[index].Open == 0 {
				s.candles[index].Open = currentPrice
			}
			if s.candles[index].High == 0 {
				s.candles[index].High = currentPrice
			}
			if s.candles[index].Low == 0 {
				s.candles[index].Low = currentPrice
			}
			if currentPrice <= s.candles[index].Low {
				s.candles[index].Low = currentPrice
			}
			if currentPrice >= s.candles[index].High {
				s.candles[index].High = currentPrice
			}
			s.candles[index].Close = currentPrice
		}
		if data.Volume != nil {
			s.candles[index].Volume += volume
		}
		if s.candles[index].Timestamp == 0 {
			s.candles[index].Timestamp = utils.GetTimestampWith0Seconds(s.currentExecutionTime)
		}
	}

}

func (s *Strategy) initCandles() {
	s.candles = nil
	s.candles = []*Candle{&Candle{}}

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
	lastCandle := s.getLastCandle()
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
		s.Logger.Error("Error when writting the last candle in the csv file -> " + err.Error())
	}
	s.csvFileMtx.Unlock()
}

func (s *Strategy) getLastCandle() *Candle {
	return s.candles[len(s.candles)-1]
}
