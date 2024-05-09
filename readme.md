# Query Benchmarking Tool for TimescaleDB

CLI tool used to benchmark `SELECT` query performance against a TimescaleDB instance.

## How To Use

### Setup

Once your TimescaleDB instance is ready ([link to official Getting Started docs](https://docs.timescale.com/getting-started/latest/)), you can use `make` to build & run a Docker container which connects to the instance & sets up schema & data.

1) Edit `Dockerfile-setup` & fill in connection details.

	> NOTE: Storing credentials in a plaintext file is **not** something I would do in production code. Ideally I would encrypt & store this data in a secrets manager, use more appropriate priviledges or a service account, & other security best practices.

2) Run `$ make setup`

## TODO / Questions

[ ] Create `homework` db. Getting error on create:
	```
	ERROR:  tsdb_admin: database homework is not an allowed database name
	HINT:  Contact your administrator to configure the "tsdb_admin.allowed_databases"
	```

[ ] Are ranges inclusive or exclusive on either end?
	- DOC THIS ANSWER IN HERE! It's important to know that when using the tool.

[ ] Ok to require Docker on client that is used as part of project setup? (To avoid installing Postgres locally)