package main

import (
	"fmt"
	"flag"
)


// type DatabaseConfig struct {
// 	Username string
// 	Password string
// 	Host 	 string
// 	DbName	 string
// 	Args 	 string
// }

// App modes:
//  "file" (default): Reads from input file arg
//  "interactive": Reads from user input, REPL-style
type Mode string
const (
	MODE_FILE			Mode = "file"
	MODE_INTERACTIVE 	Mode = "interactive"
)

type InputArgs struct {
	mode			Mode
	concurrency 	uint
	csvQueryFile	string // File path
	outputQueryResults	bool
}

func getInputArgs() *InputArgs {
	modeArg := flag.String("mode", "file", "App mode (file or interactive)")
	concurrencyArg := flag.Uint("concurrency", 1, "Number of queries that can run in parallel")
	csvQueryFileArg := flag.String("csvQueryFile", "./data/query_params.csv", "Path to query CSV file")
	outputQueryResultsArg := flag.Bool("outputQueryResults", false, "Boolean flag (true or false) indicating if query results should be output. NOTE: This can result in large output!")

	flag.Parse()

	return &InputArgs{
		mode: Mode(*modeArg),
		concurrency: *concurrencyArg,
		csvQueryFile: *csvQueryFileArg,
		outputQueryResults: *outputQueryResultsArg,
	}
}

func main() {
	inputArgs := getInputArgs()

	queryTool := NewQueryTool(inputArgs.concurrency, inputArgs.outputQueryResults)
	mode := Mode(inputArgs.mode)

	if mode == MODE_FILE {
		queryTool.RunWithCsvFile(inputArgs.csvQueryFile)
	} else if mode == MODE_INTERACTIVE {
		queryTool.RunWithManualInput(inputArgs.csvQueryFile)
	} else {
		panic(fmt.Sprintf("Unknown mode %s", mode))
	}
}

// TODO: Load env vars from .env
