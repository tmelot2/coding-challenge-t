#!/bin/bash

# Builds & runs a Docker container that uses a Postgres client to setup schema + data
# on Postgres server.

set -euo pipefail

setupName="query_tool_db_setup"

docker build -f ./Dockerfile-setup -t $setupName .
docker run -it --rm --env-file .env --name $setupName $setupName
