package parser

import (
	"encoding/csv"
	"os"
	"strconv"
)

func ReadCSV(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	return records, err
}

func ParseLines(lines [][]string) [][]float64 {
	data := make([][]float64, len(lines))
	for i, line := range lines {
		for _, strNum := range line {
			floatNum, _ := strconv.ParseFloat(strNum, 64)
			data[i] = append(data[i], floatNum)
		}
	}
	return data
}
