package main

import (
	"TradingBot/src/strategies/markets/interfaces"
	"TradingBot/src/types"
	"encoding/csv"
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

	simulatorAPI := simulator.CreateAPIServiceInstance()
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

		// Here, before calling OnNewCandle, need to check the current status
		// basically, if the price reached a limit or a stop order, and act accordingly
		// if has reached an limit or stop order: open position if didn't have, or close position if it had
		// in case of closing a position, update the balance

		// orders := simulatorAPI.GetOrders()
		// positions := simulatorAPI.GetPositions()

		strat.Parent().SetCurrentBrokerQuote(&api.Quote{
			Ask:    float32(candle.Close),
			Bid:    float32(candle.Close),
			Price:  float32(candle.Close),
			Volume: 0,
		})

		updateAPIData(&APIData, simulatorAPI)

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

func updateAPIData(APIData api.DataInterface, simulatorAPI api.Interface) {
	orders, _ := simulatorAPI.GetOrders()
	positions, _ := simulatorAPI.GetPositions()
	state, _ := simulatorAPI.GetState()

	APIData.SetOrders(orders)
	APIData.SetPositions(positions)
	APIData.SetState(state)
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
