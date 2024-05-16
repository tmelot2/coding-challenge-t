#!/bin/bash

# Runs the query tool. Input args are handled by the tool (see readme).

set -eo pipefail

go run \
	main.go \
	cpuBenchmark.go \
	database.go \
	queryTool.go \
	queue.go \
		"$@"
