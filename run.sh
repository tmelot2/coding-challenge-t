#!/bin/bash

# Runs the query tool. Input args are handled by the tool (see readme).

set -eo pipefail

# Postgres
go get github.com/lib/pq
# CPU cycle benchmarking
go get github.com/dterei/gotsc
# Large int output formatting
go get golang.org/x/text/language
go get golang.org/x/text/message

go run \
	main.go \
	cpuBenchmark.go \
	database.go \
	queryTool.go \
	queue.go \
		"$@"
