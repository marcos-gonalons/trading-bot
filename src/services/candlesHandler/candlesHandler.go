package candlesHandler

import (
	"TradingBot/src/services/logger"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper"
)

// Service ...
type Service struct {
	Logger    logger.Interface
	Symbol    string
	Timeframe types.Timeframe

	candles     []*types.Candle
	csvFileName string
	csvFileMtx  sync.Mutex
}

// InitCandles ...
func (s *Service) InitCandles() {
	s.candles = nil
	s.candles = []*types.Candle{&types.Candle{}}

	s.csvFileName = s.getCSVFileName()
	csvFile, err := os.OpenFile(s.csvFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	defer csvFile.Close()
	if err != nil {
		s.csvFileMtx.Lock()
		csvFile, err = os.Create(s.csvFileName)
		defer csvFile.Close()
		if err != nil {
			s.csvFileMtx.Unlock()
			s.Logger.Error("Error while creating the csv file -> " + err.Error())
		} else {
			csvFile.Write([]byte("Time,Open,High,Low,Close,Volume\n"))
			s.csvFileMtx.Unlock()
		}
	}

}

// UpdateCandles ...
func (s *Service) UpdateCandles(
	data *tradingviewsocket.QuoteData,
	currentExecutionTime time.Time,
	previousExecutionTime time.Time,
	lastVolume float64,
) {
	currentMinutes := currentExecutionTime.Format("04")
	previousMinutes := previousExecutionTime.Format("04")

	var currentPrice float64
	if data.Price != nil {
		currentPrice = *data.Price
	} else {
		currentPrice = 0
	}

	var volume float64
	if data.Volume != nil {
		volume = *data.Volume - lastVolume
		if volume < 0 {
			volume = *data.Volume
		}
	} else {
		volume = 0
	}

	// Todo: Depending on the timeframe, I need to change this condition
	// For 1m it can stay like this, for 15m, for example, it has to be done differently.
	if currentMinutes != previousMinutes {
		s.updateCSVWithLastCandle()
		lastCandle, _ := json.Marshal(s.GetLastCandle())
		s.Logger.Log("Adding new candle to the candles array -> " + string(lastCandle))
		s.candles = append(s.candles, &types.Candle{
			Open:      s.GetLastCandle().Close,
			Low:       s.GetLastCandle().Close,
			High:      s.GetLastCandle().Close,
			Close:     s.GetLastCandle().Close,
			Volume:    volume,
			Timestamp: utils.GetTimestampWith0Seconds(currentExecutionTime),
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
			s.candles[index].Timestamp = utils.GetTimestampWith0Seconds(currentExecutionTime)
		}
	}
}

// GetCandles ...
func (s *Service) GetCandles() []*types.Candle {
	return s.candles
}

// GetLastCandle ...
func (s *Service) GetLastCandle() *types.Candle {
	return s.candles[len(s.candles)-1]
}

func (s *Service) updateCSVWithLastCandle() {
	lastCandle := s.GetLastCandle()
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

func (s *Service) getCSVFileName() string {
	now := time.Now()
	return s.Symbol + "-" + s.Timeframe.Unit + strconv.Itoa(int(s.Timeframe.Value)) + "-" + now.Format("2006-01-02") + "-candles.csv"
}
