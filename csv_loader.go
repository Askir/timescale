package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

// LoadQueryParams reads a CSV file and returns a slice of QueryParams
func LoadQueryParams(filename string) ([]QueryParams, error) {
	records, err := readCSV(filename)
	if err != nil {
		return nil, err
	}

	return parseRecords(records)
}

func readCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

func parseRecords(records [][]string) ([]QueryParams, error) {
	var params []QueryParams
	for i, record := range records[1:] { // Skip header row
		param, err := parseRecord(record, i+2)
		if err != nil {
			return nil, err
		}
		params = append(params, param)
	}
	return params, nil
}

func parseRecord(record []string, rowNum int) (QueryParams, error) {
	startTime, err := parseTime(record[1], "start time", rowNum)
	if err != nil {
		return QueryParams{}, err
	}
	endTime, err := parseTime(record[2], "end time", rowNum)
	if err != nil {
		return QueryParams{}, err
	}
	return QueryParams{
		Hostname:  record[0],
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

func parseTime(timeStr, fieldName string, rowNum int) (time.Time, error) {
	parsedTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing %s in row %d: %v", fieldName, rowNum, err)
	}
	return parsedTime, nil
}
