#!/bin/bash

set -euo pipefail

setupName="query_tool_db_setup"

docker build -f ./Dockerfile-setup -t $setupName .
docker run -it --rm --name $setupName $setupName
