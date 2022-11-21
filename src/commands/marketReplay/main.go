package main

import (
	"TradingBot/src/commands/marketReplay/brokerSim"
	"TradingBot/src/markets"
	"TradingBot/src/services"
	"TradingBot/src/types"
	"TradingBot/src/utils"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"TradingBot/src/services/api"
	"TradingBot/src/services/api/simulator"
	"TradingBot/src/services/positionSize"

	"TradingBot/src/manager"
)

const initialBalance = float64(5000)

func main() {
	container := services.GetServicesContainer()
	container.Initialize(false)

	marketName := getMarketName()

	market := getMarketInstance(
		&manager.Manager{
			ServicesContainer: container,
		},
		marketName,
	)

	candlesFile := getCSVFile(market)

	csvLines, err := csv.NewReader(candlesFile).ReadAll()
	if err != nil {
		panic("Error while reading the .csv file -> " + err.Error())
	}

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

	if getReplayType() == "single" {
		candlesLoop(csvLines, market, container, simulatorAPI, true)
		PrintTrades(simulatorAPI.GetTrades())
		state, _ := simulatorAPI.GetState()
		fmt.Println("Profits -> ", state.Balance-initialBalance)
	} else {
		combinations, combinationsLength := GetCombinations(market.GetMarketData().MinPositionSize)
		candlesLoopWithCombinations(csvLines, market, container, simulatorAPI, combinations, combinationsLength)
	}

}

func getCSVFile(market markets.MarketInterface) *os.File {
	directory := "./.candles-csv/"
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
		if file.Name() != market.GetMarketData().CandlesFileName {
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
	if len(os.Args) < 2 {
		panic("market not specified")
	}

	return os.Args[1]
}

func getReplayType() string {
	if len(os.Args) < 3 {
		return "single"
	}

	t := os.Args[2]
	if t != "single" && t != "combo" {
		return "single"
	}

	return t
}

func candlesLoop(
	csvLines [][]string,
	market markets.MarketInterface,
	container *services.Container,
	simulatorAPI api.Interface,
	printProgress bool,
) {
	for i, line := range csvLines {
		candle := getCandleObject(line)
		market.GetCandlesHandler().AddNewCandle(candle)

		container.IndicatorsService.AddIndicators(market.GetCandlesHandler().GetCompletedCandles(), true)

		market.SetCurrentBrokerQuote(&api.Quote{
			Ask:    candle.Close,
			Bid:    candle.Close,
			Price:  candle.Close,
			Volume: 0,
		})

		SetPositionSizeStrategy(market)

		brokerSim.OnNewCandle(
			container.APIData,
			container.API,
			market,
		)

		market.OnNewCandle()

		if printProgress {
			// todo: print only every X iterations, otherwise it's prints too much
			fmt.Println(float64(i+1) * 100.0 / float64(len(csvLines)))
		}
	}
}

func SetPositionSizeStrategy(market markets.MarketInterface) {
	if market.GetMarketData().LongSetupParams != nil {
		market.GetMarketData().LongSetupParams.PositionSizeStrategy = positionSize.BASED_ON_MIN_SIZE
	}
	if market.GetMarketData().ShortSetupParams != nil {
		market.GetMarketData().ShortSetupParams.PositionSizeStrategy = positionSize.BASED_ON_MIN_SIZE
	}
}

func PrintTrades(trades []*api.Trade) {
	fmt.Println("Total trades -> ", len(trades))

	for _, trade := range trades {
		fmt.Println(
			trade.Side,
			" | ",
			utils.FloatToString(trade.Size, 0),
			" | ",
			utils.FloatToString(trade.InitialPrice, 5),
			" | ",
			utils.FloatToString(trade.FinalPrice, 5),
			" | ",
			utils.FloatToString(trade.Result, 2),
			" | ",
			trade.OpenedAt.Format("02/01/2006 15:04:05"),
			" | ",
			trade.ClosedAt.Format("02/01/2006 15:04:05"),
		)
	}
}
