package main

import (
	"flag"
	"fmt"
)

// type DatabaseConfig struct {
// 	Username string
// 	Password string
// 	Host 	 string
// 	DbName	 string
// 	Args 	 string
// }

type InputArgs struct {
	concurrency 	uint
	csvQueryFile	string // File path
	outputQueryResults	bool
}

func getInputArgs() *InputArgs {
	concurrencyArg := flag.Uint("concurrency", 1, "Number of queries that can run in parallel")
	csvQueryFileArg := flag.String("csvQueryFile", "./data/query_params.csv", "Path to query CSV file")
	outputQueryResultsArg := flag.String("outputQueryResults", false, "Flag indicating if query results should be output. NOTE: This can result in large output!")

	flag.Parse()

	return &InputArgs{
		concurrency: *concurrencyArg,
		csvQueryFile: *csvQueryFileArg,
	}
}

func main() {
	inputArgs := getInputArgs()

	queryTool := NewQueryTool(inputArgs.concurrency)
	queryTool.RunWithCsvFile(inputArgs.csvQueryFile)
}

// TODO: Load env vars from .env
