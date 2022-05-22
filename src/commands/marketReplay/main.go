package main

import (
	"TradingBot/src/commands/marketReplay/brokerSim"
	"TradingBot/src/strategies/markets/interfaces"
	"TradingBot/src/types"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"

	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/api/simulator"
	logger "TradingBot/src/services/logger/nullLogger"
	"TradingBot/src/strategies"
)

func main() {
	brokerSim := &brokerSim.BrokerSim{}
	candlesFile := getCSVFile()

	csvLines, err := csv.NewReader(candlesFile).ReadAll()
	if err != nil {
		panic("Error while reading the .csv file -> " + err.Error())
	}

	simulatorAPI := simulator.CreateAPIServiceInstance()
	simulatorAPI.SetState(&api.State{
		Balance:      1000,
		UnrealizedPL: 0,
		Equity:       1000,
	})

	APIData := api.Data{}
	strat := getMarketInstance(
		simulatorAPI,
		&APIData,
		getMarketName(),
	)
	strat.Parent().SetEurExchangeRate(.85)

	for i, line := range csvLines {
		if i == 0 {
			continue
		}

		candle := getCandleObject(line)
		strat.Parent().GetCandlesHandler().AddNewCandle(candle)

		strat.Parent().SetCurrentBrokerQuote(&api.Quote{
			Ask:    float32(candle.Close),
			Bid:    float32(candle.Close),
			Price:  float32(candle.Close),
			Volume: 0,
		})

		brokerSim.OnNewCandle(&APIData, simulatorAPI, strat)

		strat.OnNewCandle()
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

func getMarketInstance(
	simulatorAPI api.Interface,
	APIData api.DataInterface,
	marketName string,
) interfaces.MarketInterface {
	apiRetryFacade := &retryFacade.APIFacade{
		API:    simulatorAPI,
		Logger: logger.GetInstance(),
	}
	handler := &strategies.Handler{
		Logger:         logger.GetInstance(),
		API:            simulatorAPI,
		APIRetryFacade: apiRetryFacade,
		APIData:        APIData,
	}

	for _, market := range handler.GetMarkets() {
		if market.Parent().GetSymbol().BrokerAPIName == marketName {
			return market
		}
	}

	panic("invalid market name")
}

func getMarketName() string {
	if len(os.Args) != 2 {
		panic("market not specified")
	}

	return os.Args[1]
}
