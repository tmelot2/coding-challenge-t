package main

import (
	"flag"
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
	outputQueryResultsArg := flag.Bool("outputQueryResults", false, "Boolean flag (true or false) indicating if query results should be output. NOTE: This can result in large output!")

	flag.Parse()

	return &InputArgs{
		concurrency: *concurrencyArg,
		csvQueryFile: *csvQueryFileArg,
		outputQueryResults: *outputQueryResultsArg,
	}
}

func main() {
	inputArgs := getInputArgs()

	queryTool := NewQueryTool(inputArgs.concurrency, inputArgs.outputQueryResults)
	queryTool.RunWithCsvFile(inputArgs.csvQueryFile)
}

// TODO: Load env vars from .env
