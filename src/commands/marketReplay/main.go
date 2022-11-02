package main

import (
	"TradingBot/src/commands/marketReplay/brokerSim"
	"TradingBot/src/markets"
	"TradingBot/src/services"
	"TradingBot/src/types"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"TradingBot/src/services/api"
	"TradingBot/src/services/api/simulator"

	"TradingBot/src/manager"
)

const initialBalance = float64(5000)

func main() {
	candlesFile := getCSVFile()

	csvLines, err := csv.NewReader(candlesFile).ReadAll()
	if err != nil {
		panic("Error while reading the .csv file -> " + err.Error())
	}

	container := services.GetServicesContainer()
	container.Initialize(false)

	simulatorAPI := simulator.CreateAPIServiceInstance(
		&api.Credentials{},
		container.HttpClient,
		container.Logger,
	)

	container.SetAPI(simulatorAPI)

	simulatorAPI.SetState(&api.State{
		Balance:      initialBalance,
		UnrealizedPL: 0,
		Equity:       initialBalance,
	})

	market := getMarketInstance(
		&manager.Manager{
			ServicesContainer: container,
		},
		getMarketName(),
	)

	combinations := GetCombinations()
	if combinations == nil {
		candlesLoop(csvLines, market, container, simulatorAPI)
	} else {
		candlesLoopWithCombinations(csvLines, market, container, simulatorAPI, combinations)
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
	manager *manager.Manager,
	marketName string,
) markets.MarketInterface {
	for _, market := range manager.GetMarkets() {
		if market.GetMarketData().BrokerAPIName == marketName {
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

func candlesLoop(
	csvLines [][]string,
	market markets.MarketInterface,
	container *services.Container,
	simulatorAPI api.Interface,
) {
	for _, line := range csvLines {
		candle := getCandleObject(line)
		market.GetCandlesHandler().AddNewCandle(candle)

		container.IndicatorsService.AddIndicators(market.GetCandlesHandler().GetCandles(), true)

		market.SetCurrentBrokerQuote(&api.Quote{
			Ask:    candle.Close,
			Bid:    candle.Close,
			Price:  candle.Close,
			Volume: 0,
		})

		brokerSim.OnNewCandle(
			container.APIData,
			container.API,
			market,
		)

		market.OnNewCandle()

		// fmt.Println(float64(i) * 100.0 / float64(len(csvLines)))
	}

	fmt.Println("Total trades -> ", simulatorAPI.GetTrades())
	state, _ := simulatorAPI.GetState()
	fmt.Println("Profits -> ", state.Balance-initialBalance)
}
