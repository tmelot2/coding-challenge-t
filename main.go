package main

// TODO: Rename this file

import (
	// "fmt"
	"io/ioutil"
)

// type DatabaseConfig struct {
// 	Username string
// 	Password string
// 	Host 	 string
// 	DbName	 string
// 	Args 	 string
// }

// Returns file contents as a string
func readFile(filePath string) string {
	query, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	queryStr := string(query)
	return queryStr
}

func main() {
	csvFilePath := "query_params_tiny.csv"

	queryTool := newQueryTool()
	queryTool.RunWithCsvFile(csvFilePath)
}
