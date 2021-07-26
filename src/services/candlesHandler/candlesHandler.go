package candlesHandler

import (
	"TradingBot/src/services/logger"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"

	tradingviewsocket "github.com/marcos-gonalons/tradingview-scraper/v2"
)

// Service ...
type Service struct {
	Logger    logger.Interface
	Symbol    *types.Symbol
	Timeframe types.Timeframe

	candles     []*types.Candle
	csvFileName string
	csvFileMtx  sync.Mutex
}

// InitCandles ...
func (s *Service) InitCandles(currentExecutionTime time.Time, fileName string) {
	if fileName != "" {
		s.csvFileName = fileName
		s.initCandlesFromFile(currentExecutionTime)
	} else {
		s.candles = []*types.Candle{{
			Open:      0,
			High:      0,
			Low:       0,
			Close:     0,
			Volume:    0,
			Timestamp: utils.GetTimestamp(currentExecutionTime, s.getTimeLayout()),
		}}
		now := time.Now()
		s.csvFileName = s.Symbol.BrokerAPIName + "-" + s.Timeframe.Unit + strconv.Itoa(int(s.Timeframe.Value)) + "-" + now.Format("2006-01-02") + "-candles.csv"
		s.createNewCSV()
	}
}

// UpdateCandles ...
func (s *Service) UpdateCandles(
	data *tradingviewsocket.QuoteData,
	currentExecutionTime time.Time,
	lastVolume float64,
) {
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
			volume = 0
		}
	} else {
		volume = 0
	}

	if s.shouldAddNewCandle(currentExecutionTime) {
		s.updateCSVWithLastCandle()
		lastCandle, _ := json.Marshal(s.GetLastCandle())
		s.Logger.Log("Adding new candle to the candles array -> " + string(lastCandle))
		s.candles = append(s.candles, &types.Candle{
			Open:      s.GetLastCandle().Close,
			Low:       s.GetLastCandle().Close,
			High:      s.GetLastCandle().Close,
			Close:     s.GetLastCandle().Close,
			Volume:    volume,
			Timestamp: utils.GetTimestamp(currentExecutionTime, s.getTimeLayout()),
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
			s.candles[index].Timestamp = utils.GetTimestamp(currentExecutionTime, s.getTimeLayout())
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
	if err != nil {
		s.Logger.Error("Error when opening the csv file -> " + err.Error())
	}
	defer csvFile.Close()

	s.csvFileMtx.Lock()
	_, err = csvFile.Write([]byte(row))
	if err != nil {
		s.Logger.Error("Error when writting the last candle in the csv file -> " + err.Error())
	}
	s.csvFileMtx.Unlock()
}

func (s *Service) shouldAddNewCandle(currentExecutionTime time.Time) bool {
	var multiplier uint
	var currentTimestamp int64
	var candleDurationInSeconds uint

	if s.Timeframe.Unit == "m" {
		multiplier = 60
	}
	if s.Timeframe.Unit == "h" {
		multiplier = 60 * 60
	}
	if s.Timeframe.Unit == "d" {
		multiplier = 60 * 60 * 24
	}

	candleDurationInSeconds = s.Timeframe.Value * multiplier
	currentTimestamp = utils.GetTimestamp(currentExecutionTime, s.getTimeLayout())

	return currentTimestamp-s.GetLastCandle().Timestamp >= int64(candleDurationInSeconds)
}

func (s *Service) createNewCSV() {
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

func (s *Service) initCandlesFromFile(currentExecutionTime time.Time) {
	csvFile, err := os.OpenFile(s.csvFileName, os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		panic("Error while opening the .csv file -> " + err.Error())
	}

	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		panic("Error while reading the .csv file -> " + err.Error())
	}

	for index, line := range csvLines {
		s.candles = append(s.candles, &types.Candle{
			Timestamp: 0,
			Open:      s.getAsFloat64(line[1], index),
			High:      s.getAsFloat64(line[2], index),
			Low:       s.getAsFloat64(line[3], index),
			Close:     s.getAsFloat64(line[4], index),
			Volume:    s.getAsFloat64(line[5], index),
		})
	}

	s.candles = append(s.candles, &types.Candle{
		Open:      s.GetLastCandle().Close,
		Low:       s.GetLastCandle().Close,
		High:      s.GetLastCandle().Close,
		Close:     s.GetLastCandle().Close,
		Volume:    0,
		Timestamp: utils.GetTimestamp(currentExecutionTime, s.getTimeLayout()),
	})
}

func (s *Service) getAsFloat64(v string, index int) float64 {
	r, e := strconv.ParseFloat(v, 64)
	if e != nil {
		panic("Error while reading csv line at " + strconv.Itoa(index) + e.Error())
	}
	return r
}

func (s *Service) getTimeLayout() string {
	if s.Timeframe.Unit == "m" {
		return "15:04:00"
	}

	if s.Timeframe.Unit == "h" {
		return "15:00:00"
	}

	if s.Timeframe.Unit == "d" {
		return "00:00:00"
	}

	return ""
}
