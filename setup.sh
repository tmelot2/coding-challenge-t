#!/bin/bash

set -euo pipefail

connectionStr="postgres://$DB_USER:$DB_PASS@$DB_URL/$DB_DATABASE?$DB_OPTIONS"

# Setup schema
psql $connectionStr -f cpu_usage.sql

# Insert data from CSV
psql $connectionStr -c "\COPY cpu_usage FROM cpu_usage.csv CSV HEADER"
