#!/bin/bash

# Runs the query tool. Input args:
#
# $1 concurrency
# $2 query file path

set -eo pipefail

# # Default to blank
# concurrency="${1:-}"
# queryFilePath="${2:-}"

go run main.go queryTool.go queue.go "$@"
