#!/bin/bash

# Runs the query tool. Input args are handled by the tool (see readme).

set -eo pipefail

go run main.go queryTool.go queue.go database.go "$@"
