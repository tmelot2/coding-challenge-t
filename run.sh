#!/bin/bash

# Runs the query tool.
#
# Optional input args (handled by Go):
#	mode (string): Read input from file or REPL
#	concurrency (int): Number of queries to run in parallel
#	csvQueryFile (string): Path to CSV query file
#	outputQueryResults (bool): Flag indicating if query results should be output
#
# See readme or main.go for more info on input args.
#
# Example usages:
# $ bash run.sh
# $ bash run.sh -concurrency=10 -csvQueryFile=myCsvFile.csv
# $ bash run.sh -outputQueryResults=true
# $ bash run.sh -mode=interactive -concurrency=10 csvQueryFile=myCsvFile.csv

set -eo pipefail

go run main.go queryTool.go queue.go "$@"
