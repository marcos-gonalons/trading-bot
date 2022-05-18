package main

import (
	"TradingBot/src/strategies/tickers/interfaces"
	"TradingBot/src/types"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/api/simulator"
	"TradingBot/src/services/logger"
	"TradingBot/src/strategies"
)

func main() {
	candlesFile := getCSVFile()

	csvLines, err := csv.NewReader(candlesFile).ReadAll()
	if err != nil {
		panic("Error while reading the .csv file -> " + err.Error())
	}

	strat := getStrat()
	for i, line := range csvLines {
		if i == 0 {
			continue
		}
		candle := getCandleObject(line)

		fmt.Printf("%#v", candle)
	}
}

func getCSVFile() *os.File {
	directory := "./"
	osDir, err := os.Open(directory)
	if err != nil {
		panic("Error opening the directory -> " + err.Error())
	}
	files, err := osDir.Readdir(0)
	if err != nil {
		panic("Error reading the directory -> " + err.Error())
	}

	var csvFiles []os.FileInfo
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".csv" {
			continue
		}
		csvFiles = append(csvFiles, file)
	}

	if len(csvFiles) == 0 {
		panic("Couldn't find the CSV file")
	}

	csvFile, err := os.OpenFile(directory+csvFiles[0].Name(), os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		panic("Error while opening the .csv file -> " + err.Error())
	}

	return csvFile
}

func getCandleObject(csvLine []string) (candle types.Candle) {
	timestamp, _ := strconv.ParseInt(csvLine[0], 10, 64)
	open, _ := strconv.ParseFloat(csvLine[1], 64)
	high, _ := strconv.ParseFloat(csvLine[2], 64)
	low, _ := strconv.ParseFloat(csvLine[3], 64)
	close, _ := strconv.ParseFloat(csvLine[4], 64)
	volume, _ := strconv.ParseFloat(csvLine[5], 64)

	candle.Timestamp = timestamp
	candle.Open = open
	candle.High = high
	candle.Low = low
	candle.Close = close
	candle.Volume = volume

	return
}

func getStrat() interfaces.TickerInterface {
	simulatorAPI := simulator.CreateAPIServiceInstance()

	apiRetryFacade := &retryFacade.APIFacade{
		API:    simulatorAPI,
		Logger: logger.GetInstance(),
	}
	handler := &strategies.Handler{
		Logger:         logger.GetInstance(),
		API:            simulatorAPI,
		APIRetryFacade: apiRetryFacade,
		APIData:        &api.Data{},
	}

	for _, strat := range handler.GetStrategies() {
		if strat.Parent().GetSymbol().BrokerAPIName == getSymbolName() {
			return strat
		}
	}

	panic("Invalid symbol name")
}

func getSymbolName() string {
	if len(os.Args) != 2 {
		panic("Symbol not specified")
	}

	return os.Args[1]
}
