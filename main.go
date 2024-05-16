package main

import (
	"flag"
	"fmt"
	"github.com/dterei/gotsc"
)

var benchmarker *CpuBenchmark

// App modes:
//
//	"file" (default): Reads from input file arg
//	"interactive": Reads from user input, REPL-style
type Mode string

const (
	MODE_FILE        Mode = "file"
	MODE_INTERACTIVE Mode = "interactive"
)

// All app input arguments
type InputArgs struct {
	mode               Mode
	concurrency        uint
	csvQueryFile       string // File path
	outputQueryResults bool
}

// Returns app input args or defaults
func getInputArgs() *InputArgs {
	modeArg := flag.String("mode", "file", "App mode (file or interactive)")
	concurrencyArg := flag.Uint("concurrency", 1, "Number of queries that can run in parallel")
	csvQueryFileArg := flag.String("csvQueryFile", "./data/query_params.csv", "Path to query CSV file")
	outputQueryResultsArg := flag.Bool("outputQueryResults", false, "Boolean flag (true or false) indicating if query results should be output. NOTE: This can result in large output!")

	flag.Parse()

	return &InputArgs{
		mode:               Mode(*modeArg),
		concurrency:        *concurrencyArg,
		csvQueryFile:       *csvQueryFileArg,
		outputQueryResults: *outputQueryResultsArg,
	}
}

// ////////////////////
// Main!
func main() {
	// Create CPU cycle benchmarking tool
	benchmarker = NewCpuBenchmark(256)

	// Load env file
	cyclesStart := gotsc.BenchStart()
	db := NewDatabase("./.env")
	cyclesEnd := gotsc.BenchEnd()
	benchmarker.Add(BenchmarkTypeEnvFile, cyclesEnd - cyclesStart)

	// Get input args
	cyclesStart = gotsc.BenchStart()
	inputArgs := getInputArgs()
	cyclesEnd = gotsc.BenchEnd()
	benchmarker.Add(BenchmarkTypeInputArgs, cyclesEnd - cyclesStart)

	// Create the tool
	queryTool := NewQueryTool(db, inputArgs.concurrency, inputArgs.outputQueryResults)

	// Run tool in the mode specified
	mode := Mode(inputArgs.mode)
	if mode == MODE_FILE {
		queryTool.RunWithCsvFile(inputArgs.csvQueryFile)
	} else if mode == MODE_INTERACTIVE {
		queryTool.RunWithManualInput(inputArgs.csvQueryFile)
	} else {
		panic(fmt.Sprintf("Unknown mode %s", mode))
	}

	benchmarker.Print()
}
