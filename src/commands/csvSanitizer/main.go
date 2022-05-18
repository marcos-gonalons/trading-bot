package main

import (
	"TradingBot/src/utils"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	directory := getDirectory()
	csvFiles := getCsvInDir(directory)

	for _, file := range csvFiles {
		csvFile, err := os.OpenFile(directory+file.Name(), os.O_APPEND|os.O_RDWR, os.ModeAppend)
		if err != nil {
			panic("Error while opening the .csv file -> " + err.Error())
		}

		tmpCsvFileName := "./" + utils.GetRandomString(10) + file.Name() + ".csv"
		tmpCsvFile, err := os.Create(tmpCsvFileName)
		if err != nil {
			panic("Error creating tmp file ->" + err.Error())
		}

		csvLines, err := csv.NewReader(csvFile).ReadAll()
		if err != nil {
			panic("Error while reading the .csv file -> " + err.Error())
		}

		for i, line := range csvLines {
			if i == 0 {
				continue
			}
			newLine := getTimestamp(line[0]) + "," + line[1] + "," + line[2] + "," + line[3] + "," + line[4] + ",0\n"
			tmpCsvFile.Write([]byte(newLine))
		}

		csvFile.Close()
		tmpCsvFile.Close()

		err = os.Remove(directory + file.Name())
		if err != nil {
			panic("Error while removing the csv file -> " + err.Error())
		}

		err = os.Rename(tmpCsvFileName, directory+file.Name())
		if err != nil {
			panic("Error renaming the temp csv file -> " + err.Error())
		}
	}

}

func getDirectory() string {
	if len(os.Args) != 2 {
		panic("Directory not specified")
	}

	return os.Args[1]
}

func getCsvInDir(dir string) []os.FileInfo {
	osDir, err := os.Open(dir)
	if err != nil {
		panic("Error opening the directory " + dir + " -> " + err.Error())
	}
	files, err := osDir.Readdir(0)
	if err != nil {
		panic("Error reading the directory " + dir + " -> " + err.Error())
	}

	var csvFiles []os.FileInfo
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".csv" {
			continue
		}
		csvFiles = append(csvFiles, file)
	}

	return csvFiles
}

// Example of v: 05.11.2021 21:00:00.000
func getTimestamp(v string) string {
	d, err := time.Parse("02.01.2006 15:04:05 MST-0700", v)
	if err != nil {
		panic("Error parsing date -> " + err.Error())
	}

	return strconv.Itoa(int(d.Unix()))
}
