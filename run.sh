#!/bin/bash

# Runs the query tool

set -euo pipefail

go run main.go queryTool.go queue.go
