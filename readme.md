# Query Benchmarking Tool for TimescaleDB

CLI tool used to benchmark `SELECT` query performance against a TimescaleDB instance.


## Requirements

- An existing TimescaleDB instance.
- Local install of Docker to run the database setup container.
- Go
- Bash (or compatible)

Project was developed on MacOS 13.6.6, Go 1.22.1, Docker 26.0.0, zsh 5.9.


## Setup

This tool uses Docker to run a Postgres client that connects to the db & sets up the schema + data.

1) Start TimescaleDB instance ([link to official Getting Started docs](https://docs.timescale.com/getting-started/latest/)).

2) Edit `Dockerfile-setup` & fill in connection details.

	> TODO: Prolly moving those values to .env file.

	> NOTE: Storing credentials in a plaintext file is **not** something I would do in production code. Ideally I would encrypt & store this data in a secrets manager, use more appropriate priviledges or a service account, & other security best practices.

3) Run `$ make setup`


## Run

This tool supports 2 modes:


### Mode 1: Load queries from CSV file.

1) `$ make run`

> TODO: More, input file, concurrency

### Mode 2: Interactive via stdin
> TODO: More, concurrency


## Todo & Questions

[ ] Race condition

[ ] TODO: BUCKETING BY HASH OF HOSTNAME REDUCES PARALLELISM BECAUSE WE ONLY GET AS MANY BUCKETS AS UNIQUE HOSTNAMES
	- Therefore, host 0 to 9 = 10 hosts, so 10 buckets. Repeat hosts read from the file bucket behind previous requests, so they're stuck waiting.
	- It's like we actually need "queries per second" or something

[ ] "Total time" doesn't really make sense: Concurrency 10 yields 8 queries that ran for 1.7s, even though they ran at the same time
	- Therefore, average against the sum doesn't make sense. Should avg be something else?

[ ] Create `homework` db. Getting error on create:
	```
	ERROR:  tsdb_admin: database homework is not an allowed database name
	HINT:  Contact your administrator to configure the "tsdb_admin.allowed_databases"
	```

[ ] Are ranges inclusive or exclusive on either end?
	- DOC THIS ANSWER IN HERE! It's important to know that when using the tool.

[ ] Is it correct that we don't output query results, just benchmark results?

[ ] Ok to require Docker on client that is used as part of project setup? (To avoid installing Postgres locally)