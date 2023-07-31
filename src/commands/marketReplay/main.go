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
	replayType := getReplayType()
	side := getSide()

	market := getMarketInstance(
		&manager.Manager{
			ServicesContainer: container,
		},
		marketName,
	)
	SetPositionSizeStrategy(market)

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

	////////////////////////////////////////////////////
	strategyParamsFunc := market.SetRangesStrategyParams
	////////////////////////////////////////////////////

	if replayType == SINGLE_TYPE {
		setStrategyData(side, market, strategyParamsFunc)

		candlesLoop(csvLines, market, container, simulatorAPI, false)
		PrintTrades(simulatorAPI.GetTrades())
		state, _ := simulatorAPI.GetState()
		fmt.Println("Profits -> ", state.Balance-initialBalance)
	} else {
		combinations, combinationsLength := GetCombinations(market.GetMarketData().MinPositionSize)
		candlesLoopWithCombinations(csvLines, market, container, simulatorAPI, combinations, combinationsLength, side, strategyParamsFunc)
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

func getReplayType() ReplayType {
	var t = os.Args[2]
	switch t {
	case "single":
		return SINGLE_TYPE
	case "combo":
		return COMBO_TYPE
	}

	panic("Invalid replay type. Only allowed single or combo.")
}

func getSide() Side {
	var s = os.Args[3]
	switch s {
	case "longs":
		return LONGS_SIDE
	case "shorts":
		return SHORTS_SIDE
	}

	panic("Invalid side. Only allowed longs or shorts.")
}

func setStrategyData(
	side Side,
	market markets.MarketInterface,
	strategyParamsFunc func(longs *types.MarketStrategyParams, shorts *types.MarketStrategyParams),
) {
	if side == LONGS_SIDE {
		strategyParamsFunc(market.GetMarketData().RangesSetup.LongSetupParams, nil)
	} else if side == SHORTS_SIDE {
		strategyParamsFunc(nil, market.GetMarketData().RangesSetup.ShortSetupParams)
	}
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
	if market.GetMarketData().EmaCrossoverSetup.LongSetupParams != nil {
		market.GetMarketData().EmaCrossoverSetup.LongSetupParams.PositionSizeStrategy = positionSize.BASED_ON_MIN_SIZE
	}
	if market.GetMarketData().EmaCrossoverSetup.ShortSetupParams != nil {
		market.GetMarketData().EmaCrossoverSetup.ShortSetupParams.PositionSizeStrategy = positionSize.BASED_ON_MIN_SIZE
	}

	if market.GetMarketData().RangesSetup.LongSetupParams != nil {
		market.GetMarketData().RangesSetup.LongSetupParams.PositionSizeStrategy = positionSize.BASED_ON_MIN_SIZE
	}
	if market.GetMarketData().RangesSetup.ShortSetupParams != nil {
		market.GetMarketData().RangesSetup.ShortSetupParams.PositionSizeStrategy = positionSize.BASED_ON_MIN_SIZE
	}
}

func PrintTrades(trades []*api.Trade) {
	csvString := "Trade Number,Start Date,End Date,Start Price,End Price,Result\n"
	for index, trade := range trades {
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

		csvString = csvString +
			strconv.Itoa(index+1) + "," +
			trade.OpenedAt.Format("02/01/2006 15:04") + "," +
			trade.ClosedAt.Format("02/01/2006 15:04") + "," +
			utils.FloatToString(trade.InitialPrice, 5) + "," +
			utils.FloatToString(trade.FinalPrice, 5) + "," +
			utils.FloatToString(trade.Result, 2) + "\n"
	}
	fmt.Println("Total trades -> ", len(trades))
	fmt.Println("\n\n", csvString)
}
