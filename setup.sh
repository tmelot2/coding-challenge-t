#!/bin/bash

set -euo pipefail

connectionStr="postgres://$DB_USER:$DB_PASS@aqadz0sy32.tzug8uusr7.tsdb.cloud.timescale.com:39894/tsdb?sslmode=require"

# Setup db
psql $connectionStr -f cpu_usage.sql

# Insert data from CSV
psql $connectionStr -c "\COPY cpu_usage FROM cpu_usage.csv CSV HEADER"
