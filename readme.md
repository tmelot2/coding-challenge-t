# Query Benchmarking Tool for TimescaleDB

CLI tool used to benchmark `SELECT` query performance against a TimescaleDB instance.


## Requirements

- An existing TimescaleDB instance.
- Local install of Docker to run the database setup container.
- Make
- Go

Project was developed on MacOS 13.6.6, Go 1.22.1, Make 3.81, Docker 26.0.0.


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

[ ] Create `homework` db. Getting error on create:
	```
	ERROR:  tsdb_admin: database homework is not an allowed database name
	HINT:  Contact your administrator to configure the "tsdb_admin.allowed_databases"
	```

[ ] Are ranges inclusive or exclusive on either end?
	- DOC THIS ANSWER IN HERE! It's important to know that when using the tool.

[ ] Ok to require Docker on client that is used as part of project setup? (To avoid installing Postgres locally)