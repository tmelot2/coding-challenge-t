# Query Benchmarking Tool for TimescaleDB

CLI tool used to benchmark `SELECT` query performance against a TimescaleDB instance.


## Requirements

- An existing TimescaleDB instance.
- Local install of Docker to run the database setup container.
- Go
- Bash (or compatible)

Project was developed on MacOS 13.6.6, Go 1.22.1, Docker 26.0.0, Bash 3.2.57.


## Setup

This tool uses Docker to run a Postgres client that connects to the db & sets up the schema + data.

1) Start TimescaleDB instance ([link to official Getting Started docs](https://docs.timescale.com/getting-started/latest/)).

2) Edit `.env` & fill in connection & credential details.

	> NOTE: Do not commit credentials to the repo!

	> NOTE: Also, storing credentials in a plaintext file is **not** something I would do in production code. Ideally I would encrypt & store this data in a secrets manager, use more appropriate priviledges or a service account, & other security best practices.

3) Run `$ bash setup.sh`

4) To avoid "too many connection" type errors, configure your Timescale instance using the dashboard, Services > Your Instance:
	- Overview tab > Connect to your service
		- Use: Connection pooler, pool = Transaction pool
	- Operations tab > Compute > 8 CPU / 32 GiB Memory
	- Read scaling > Add read replica
		- Same CPU & memory, Enable connection pooler
		- Note I'm not sure if this is needed or not, haven't experimented without it.


## Run

To run the tool with default settings (see below for defaults):

`$ bash run.sh`

### Run Options

| Name | Type | Default | Description |
|----------|----------|----------|----------|
| `mode` | string | `file` | `file`: To read from input file. <br> `interactive`: REPL-like. |
| `concurrency` | int | `1` | Number of queries that can run at the same time. |
| `csvQueryFile` | string | `./data/query_params.csv` | Path to CSV file that will be translated into queries. |
| `outputQueryResults` | bool | `false` | Flag that indicates if query results should be output. |

#### Run Examples
Run with concurrency of 10:
```
$ bash run.sh -concurrency=10
```

Run with concurrency of 5, use custom query CSV file:
```
$ bash run.sh -concurrency=5 csvQueryFile=/path/to/myFile.csv
```

Run in interactive mode & print query results:
```
$ bash run.sh -mode=interactive -outputQueryResults=true
```


## CSV Query File

Format is:
```
hostname,start_time,end_time
host_000008,2017-01-01 08:59:22,2017-01-01 09:59:22
```

The query includes start time & excludes end time.


## Design Notes

### Concurrency Limitations

The requirement for workers to consistently execute queries for the same hostname imposes an upper limit on concurrency efficiency for 2 reasons:

1) When a host is queried twice, the 2nd query must wait on the 1st to finish, even if other queues are empty.
2) The # of queues actually used is limited to the # of unique hostnames.

For example, if there's 10 unique hostnames, then at best, only 10 workers will be assigned work, even if there's 100 queues available.

Efficiency could be improved by using a different bucketing method like round-robin, least recently used, or prioritizing open queues. I tested round-robin, as the code change is very simple. Results using default CSV file (200 queries, 10 unique hostnames):

- Bucketing by hostname:
	- Concurrency    5:  25s, avg 0.18s
	- Concurrency   10:  18s, avg 0.18s
	- Concurrency  100:  18s, avg 0.18s
- Round-robin
	- Concurrency  10: 9.0s, avg 0.20s
	- Concurrency  25: 7.5s, avg 0.32s
	- Concurrency  50: 7.6s, avg 0.53s
	- Concurrency 100: 7.7s, avg 1.10s
	- Concurrency 200: 7.7s, avg 2.3s

Notice that with 10 or 100 queues, hostname bucketing is the same, & round-robin is much faster, even at 10. It's also apparent that we hit a limit of 0.18s per query on average, based on the server-side load.

### Queue Decoupling

I'd prefer the Queue & Job structs to be more decoupled from the Query Tool. With a little bit of reflection work, I think that should be fairly easy to do. That way, the Queue can accept a generic Job with whatever fields & a work function.

Then the Queue could be reused with other queries, job types, or in other parts of a larger system.

### Error Handling

Because this is a small & simple tool, most error cases are simply handed by calling `panic()` with an error message.

If I were writing a library or parts of a larger system, I would refactor much of the error handling to fall in line with the standard Go practices of functions returning errors.

### Tests

Normally, I would never write any code without tests. But in this case, the project was a fairly large workload & I am short on programming time, so there are no tests 😕. This is Very Bad Practice, and I realize the irony of writing this for a job application 😄.

If time were no object, I would have begun to add unit tests during refactoring after commit `ee46b1c`, which is when I had working spaghetti code for serial queries. In the past, I've been a fan of getting external data using mocks via dependency injection. I would also add integration tests to test output of the whole system, database included.

Here's examples of unit tests I wrote for a still-in-development Go JSON parser:
- [lexer_test.go](https://github.com/tmelot2/go-json-parser/blob/dev/lexer_test.go)
- [parser_test.go](https://github.com/tmelot2/go-json-parser/blob/dev/parser_test.go)


## Usage

Anyone is free to use the code for any purpose. Happy coding!
