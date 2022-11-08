package candlesHandler

import (
	"TradingBot/src/services/candlesHandler/indicators"
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

const CandlesFolder = ".candles-csv/"

type Service struct {
	Log               func(msg string)
	MarketData        *types.MarketData
	IndicatorsBuilder indicators.MainInterface

	completedCandles []*types.Candle
	currentCandle    *types.Candle
	csvFileName      string
	csvFileMtx       sync.Mutex
}

func (s *Service) InitCandles(currentExecutionTime time.Time, fileName string) {
	if fileName != "" {
		s.csvFileName = fileName
		s.initCandlesFromFile(currentExecutionTime)

		s.Log("Adding indicators to candles ...")
		s.IndicatorsBuilder.AddIndicators(s.completedCandles, false)
	} else {
		s.currentCandle = &types.Candle{
			Open:       0,
			High:       0,
			Low:        0,
			Close:      0,
			Volume:     0,
			Indicators: types.Indicators{},
			Timestamp:  utils.GetTimestamp(currentExecutionTime, s.getTimeLayout()),
		}

		s.csvFileName = s.MarketData.BrokerAPIName + "-" + s.MarketData.Timeframe.Unit + strconv.Itoa(int(s.MarketData.Timeframe.Value)) + "-" + time.Now().Format("2006-01-02") + "-candles.csv"
		s.createCSVFile(s.csvFileName)
	}
}

func (s *Service) UpdateCandles(data *tradingviewsocket.QuoteData, lastVolume float64, onNewCandleCallback func()) {
	price, volume := s.getPriceAndVolume(data, lastVolume)

	s.Log("Updating candles data with this socket data -> " + utils.GetStringRepresentation(data))
	s.Log("Last volume is -> " + utils.FloatToString(lastVolume, 0))
	s.Log("Volume is -> " + utils.FloatToString(volume, 0))

	now := time.Now()
	timestamp := utils.GetTimestamp(now, s.getTimeLayout())
	if s.shouldCompleteCurrentCandle(now) {
		s.Log("Time to complete the current candle")
		s.completeCurrentCandle(volume, timestamp, onNewCandleCallback)
		s.Log("A new candle has been completed!")
	} else {
		s.Log("Current candle hasn't been completed yet. Updating current candle data ...")
		if s.currentCandle == nil {
			s.updateCurrentCandleWithLastCompletedCandle(volume, timestamp)
		}
		s.updateCurrentCandle(data, price, volume)
	}
}

func (s *Service) GetLastCompletedCandle() *types.Candle {
	return s.completedCandles[len(s.completedCandles)-1]
}

func (s *Service) AddNewCandle(candle types.Candle) {
	s.completedCandles = append(s.completedCandles, &candle)
}

func (s *Service) GetCompletedCandles() []*types.Candle {
	return s.completedCandles
}

func (s *Service) SetCandles(c []*types.Candle) {
	s.completedCandles = c
}

func (s *Service) RemoveOldCandles(amount uint) {
	s.completedCandles = s.completedCandles[amount:]

	tempFileName := utils.GetRandomString(10) + ".csv"
	s.createCSVFile(tempFileName)

	for _, candle := range s.completedCandles {
		s.writeRowIntoCSVFile(s.getRowForCSV(candle), tempFileName)
	}

	s.csvFileMtx.Lock()
	defer func() {
		s.csvFileMtx.Unlock()
	}()

	err := os.Remove(CandlesFolder + s.csvFileName)
	if err != nil {
		panic("Error while removing the csv file -> " + err.Error())
	}

	err = os.Rename(CandlesFolder+tempFileName, CandlesFolder+s.csvFileName)
	if err != nil {
		panic("Error renaming the temp csv file -> " + err.Error())
	}
}

func (s *Service) writeRowIntoCSVFile(row []byte, fileName string) (err error) {
	s.Log("Writing row into the csv file. Csv file is -> " + CandlesFolder + fileName)
	s.Log("Row is " + string(row))

	csvFile, err := os.OpenFile(CandlesFolder+fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return
	}
	defer csvFile.Close()

	s.csvFileMtx.Lock()
	_, err = csvFile.Write(row)
	s.csvFileMtx.Unlock()

	return
}

func (s *Service) getRowForCSV(candle *types.Candle) []byte {
	return []byte("" +
		strconv.FormatInt(candle.Timestamp, 10) + "," +
		utils.FloatToString(candle.Open, s.MarketData.PriceDecimals) + "," +
		utils.FloatToString(candle.High, s.MarketData.PriceDecimals) + "," +
		utils.FloatToString(candle.Low, s.MarketData.PriceDecimals) + "," +
		utils.FloatToString(candle.Close, s.MarketData.PriceDecimals) + "," +
		utils.FloatToString(candle.Volume, s.MarketData.PriceDecimals) + "\n")
}

func (s *Service) shouldCompleteCurrentCandle(currentExecutionTime time.Time) bool {
	if s.currentCandle == nil {
		return false
	}

	var multiplier int64
	var currentTimestamp int64
	var candleDurationInSeconds int64

	if s.MarketData.Timeframe.Unit == "m" {
		multiplier = 60
	}
	if s.MarketData.Timeframe.Unit == "h" {
		multiplier = 60 * 60
	}
	if s.MarketData.Timeframe.Unit == "d" {
		multiplier = 60 * 60 * 24
	}

	candleDurationInSeconds = int64(s.MarketData.Timeframe.Value) * multiplier
	currentTimestamp = utils.GetTimestamp(currentExecutionTime, s.getTimeLayout())

	hasEnoughTimePassed := currentTimestamp-s.currentCandle.Timestamp >= int64(candleDurationInSeconds)
	isWithinTradingHours := utils.IsWithinTradingHours(currentExecutionTime, s.MarketData.TradingHours)

	s.Log("Should the current candle be completed?")
	s.Log("currentCandle is  ->" + utils.GetStringRepresentation(s.currentCandle))
	s.Log("marketData is ->" + utils.GetStringRepresentation(s.MarketData))
	s.Log("hasEnoughTimePassed? -> " + strconv.FormatBool(hasEnoughTimePassed))
	s.Log("isWithinTradingHours? -> " + strconv.FormatBool(isWithinTradingHours))

	return hasEnoughTimePassed && isWithinTradingHours
}

func (s *Service) completeCurrentCandle(
	volume float64,
	timestamp int64,
	onNewCandleCallback func(),
) {
	err := s.writeRowIntoCSVFile(s.getRowForCSV(s.currentCandle), s.csvFileName)
	if err != nil {
		s.Log("Error when writing the current candle into the CSV file -> " + err.Error())
	}

	lastCandle, _ := json.Marshal(s.currentCandle)
	s.Log("Adding new completed candle to the completed candles array (" + s.MarketData.SocketName + ") -> " + string(lastCandle))
	s.completedCandles = append(s.completedCandles, &types.Candle{
		Open:      s.currentCandle.Open,
		High:      s.currentCandle.High,
		Low:       s.currentCandle.Low,
		Close:     s.currentCandle.Close,
		Volume:    s.currentCandle.Volume,
		Timestamp: s.currentCandle.Timestamp,
	})

	s.updateCurrentCandleWithLastCompletedCandle(volume, timestamp)

	s.Log("Current candle now is " + utils.GetStringRepresentation(s.currentCandle))

	s.IndicatorsBuilder.AddIndicators(s.completedCandles, true)
	onNewCandleCallback()
}

func (s *Service) createCSVFile(fileName string) {
	s.Log("Creating candles file -> " + fileName)

	csvFile, err := os.OpenFile(CandlesFolder+fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	defer csvFile.Close()
	if err != nil {
		s.csvFileMtx.Lock()
		csvFile, err = os.Create(CandlesFolder + fileName)
		defer csvFile.Close()
		if err != nil {
			s.csvFileMtx.Unlock()
			panic("Error while creating the csv file -> " + err.Error())
		} else {
			s.csvFileMtx.Unlock()
		}
	}
}

func (s *Service) initCandlesFromFile(currentExecutionTime time.Time) {
	s.Log("Initializing candles from file -> " + CandlesFolder + s.csvFileName)

	csvFile, err := os.OpenFile(CandlesFolder+s.csvFileName, os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		panic("Error while opening the .csv file -> " + err.Error())
	}

	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		panic("Error while reading the .csv file -> " + err.Error())
	}

	for index, line := range csvLines {
		candle := &types.Candle{
			Timestamp:  s.getAsInt64(line[0], index),
			Open:       s.getAsFloat64(line[1], index),
			High:       s.getAsFloat64(line[2], index),
			Low:        s.getAsFloat64(line[3], index),
			Close:      s.getAsFloat64(line[4], index),
			Volume:     s.getAsFloat64(line[5], index),
			Indicators: types.Indicators{},
		}
		s.completedCandles = append(s.completedCandles, candle)
	}
}

func (s *Service) getAsFloat64(v string, index int) float64 {
	r, e := strconv.ParseFloat(v, 64)
	if e != nil {
		panic("Error while reading csv line at " + strconv.Itoa(index) + e.Error())
	}
	return r
}

func (s *Service) getAsInt64(v string, index int) int64 {
	r, e := strconv.ParseInt(v, 10, 64)
	if e != nil {
		panic("Error while reading csv line at " + strconv.Itoa(index) + e.Error())
	}
	return r
}

func (s *Service) getTimeLayout() string {
	if s.MarketData.Timeframe.Unit == "m" {
		return "15:04:00"
	}

	if s.MarketData.Timeframe.Unit == "h" {
		return "15:00:00"
	}

	if s.MarketData.Timeframe.Unit == "d" {
		return "00:00:00"
	}

	return ""
}

func (s *Service) updateCurrentCandleWithLastCompletedCandle(volume float64, timestamp int64) {
	s.currentCandle = &types.Candle{
		Open:      s.GetLastCompletedCandle().Close,
		Low:       s.GetLastCompletedCandle().Close,
		High:      s.GetLastCompletedCandle().Close,
		Close:     s.GetLastCompletedCandle().Close,
		Volume:    volume,
		Timestamp: timestamp,
	}
}

func (s *Service) updateCurrentCandle(
	data *tradingviewsocket.QuoteData,
	price float64,
	volume float64,
) {
	if data.Price != nil {
		if s.currentCandle.Open == 0 {
			s.currentCandle.Open = price
		}
		if s.currentCandle.High == 0 {
			s.currentCandle.High = price
		}
		if s.currentCandle.Low == 0 {
			s.currentCandle.Low = price
		}
		if price <= s.currentCandle.Low {
			s.currentCandle.Low = price
		}
		if price >= s.currentCandle.High {
			s.currentCandle.High = price
		}
		s.currentCandle.Close = price
	}
	if data.Volume != nil {
		s.currentCandle.Volume += volume
	}
	if s.currentCandle.Timestamp == 0 {
		s.currentCandle.Timestamp = utils.GetTimestamp(time.Now(), s.getTimeLayout())
	}
}

func (s *Service) getPriceAndVolume(data *tradingviewsocket.QuoteData, lastVolume float64) (price, volume float64) {
	if data.Price != nil {
		price = *data.Price
	} else {
		price = 0
	}

	if data.Volume != nil {
		if lastVolume > 0 {
			volume = *data.Volume - lastVolume
		}
		if volume < 0 {
			volume = *data.Volume
		}
	}

	return
}
