package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"strings"
	"time"
)

import _ "github.com/lib/pq"


// Duration is sortable because it adheres to sort interface
type Duration []time.Duration
func (d Duration) Len() int       	  { return len(d) }
func (d Duration) Less(i, j int) bool { return d[i] < d[j] }
func (d Duration) Swap(i, j int)	  { d[i], d[j] = d[j], d[i] }

// QueryTool is used to run benchmarked queries against the TimescaleDB instance.
type QueryTool struct {
	multiQueue []*Queue
	queryTimes []time.Duration
}

// Returns an instance of QueryTool.
func NewQueryTool(concurrency uint) *QueryTool {
	if concurrency <= 0 {
		fmt.Printf("Concurrency <= 0 (is %d), setting to 1\n", concurrency)
		concurrency = 1
	}

	// Create multi-queue
	queues := make([]*Queue, 0, concurrency)
	for i := 0; i < int(concurrency); i++ {
		queues = append(queues, NewQueue())
	}

	// Create self
	// Initial capacity of 256 to avoid some append() data copying using the provided query CSV.
	queryTool := QueryTool{queues, make([]time.Duration, 0, 256)}

	// Start queues
	queryTool.startMultiQueue()

	return &queryTool
}

// Returns the bucket index this key hashes into for the given number of buckets.
func (queryTool *QueryTool) getQueue(key string) *Queue {
	h := fnv.New32a()
	h.Write([]byte(key))
	hash := h.Sum32()
	bucket := int(hash) % len(queryTool.multiQueue)
	fmt.Printf("queue #%d: process host %s\n", bucket, key)
	return queryTool.multiQueue[bucket]
}

// Starts all queues in the multiqueue.
func (queryTool *QueryTool) startMultiQueue() {
	for _,q := range queryTool.multiQueue {
		go q.Start()
	}
}

// Waits for all multiqueue wait groups to finish.
func (queryTool *QueryTool) waitAllMultiQueue() {
	for _,q := range queryTool.multiQueue {
		q.Wait()
	}
}

// Stops all queues in the multiqueue.
func (queryTool *QueryTool) stopMultiQueue() {
	for _,q := range queryTool.multiQueue {
		q.Stop()
	}
}

// Reads the CSV file line-by-line, turning each line into a db query that's immediately run.
// Keeps track of runtime as it goes & prints stats when finished.
func (queryTool *QueryTool) RunWithCsvFile(filePath string) {
	// Open CSV file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening CSV file:", err)
		return
	}
	defer file.Close()

	// Read & process CSV line-by-line
	scanner := bufio.NewScanner(file)
	scanner.Scan() // Ignore 1st line (it's a header)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		host, start, end := parts[0], parts[1], parts[2]

		// queryTime := queryTool.runQuery(query, start, end, host)
		// queryTool.runQuery(query, start, end, host)
		queue := queryTool.getQueue(host)
		queue.Enqueue(Job{start, end, host, queryTool.runQuery})
		queryTime := time.Duration(1*time.Second)

		// fmt.Println(queryTime)
		queryTool.queryTimes = append(queryTool.queryTimes, queryTime)

		// fmt.Printf("\n%s\n", strings.Repeat("=", 30))
	}

	// Wait on all multiqueues
	queryTool.waitAllMultiQueue()

	/*
	   BUG: This exhibits race-condition-like behavior where sometimes it errors. Even though the previous statement waits for
	   all queue wait groups to finish, calling stop() here will cause Queue.Start(): job.F(...) to error, saying nil dereference.

	   What I think that means is that the jobs channels are GC'd *BEFORE* some processing is complete, EVEN THOUGH the previous
	   line was supposed to wait for all that to finish.

	   In this case, none of those resources are used again & the program exists shortly after, so there's no real bad effect of
	   commenting this out, as far as I understand.

	   I tested running it with the -race flag, but it didn't find anything (& I think that tool may be more for when modifying
	   shared data & such).
	*/
	// Stop all queues in the multiqueue
	// queryTool.stopMultiQueue()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning file:", err)
	}

	queryTool.printQueryTimeStats()
}

// Runs the given query in the db, prints the results, & returns the runtime of the query operation.
func (queryTool *QueryTool) runQuery(start, end, host string) time.Duration {
	query := readFile("query_cpuMinMaxByMin.sql")

	conn := queryTool.getDatabaseConnection()
	defer conn.Close()

	// Run the query & time how long it takes
	queryStart := time.Now()
	fmt.Printf("%s query start %s\n", host, time.Now())

	stmt, err := conn.Prepare(query)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(start, end, host)

	fmt.Printf("%s query end   %s\n", host, time.Now())
	queryEnd := time.Now()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// Print results
	count := 0
	for rows.Next() {
		var host string
		var ts time.Time
		var cpuMin float64
		var cpuMax float64

		err := rows.Scan(&host, &ts, &cpuMin, &cpuMax)
		if err != nil {
			panic(err)
		}

		count += 1

		// fmt.Println(ts, cpuMin, cpuMax)
	}
	// fmt.Printf("%d rows\n", count)

	// Calculate runtime & return it
	elapsedTime := queryEnd.Sub(queryStart)
	return elapsedTime
}

// Connects to the db & returns the connection
// TODO: Refactor to ConnectionManager or something, & pass in a ref of that to QueryTool constructor
func (queryTool *QueryTool) getDatabaseConnection() *sql.DB {
	// TODO: Pull from .env file
	// postgres://tsdbadmin@aqadz0sy32.tzug8uusr7.tsdb.cloud.timescale.com:31633/tsdb?sslmode=require
	username := "tsdbadmin"
	password := "oda9b95cubho3tqq"
	host := "aqadz0sy32.tzug8uusr7.tsdb.cloud.timescale.com:31633"
	db := "tsdb_transaction"
	args := "sslmode=require"

	connectionString := fmt.Sprintf("postgres://%s:%s@%s/%s?%s", username, password, host, db, args)
	conn, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}

	err = conn.Ping()
	if err != nil {
		panic(err)
	}

	return conn
}

// Prints time stats related to how long the queries took
func (queryTool *QueryTool) printQueryTimeStats() {
	numQueries := len(queryTool.queryTimes)
	// Prime min & max as 0th element
	minTime := queryTool.queryTimes[0]
	maxTime := queryTool.queryTimes[0]
	var totalTime time.Duration

	// Compute min, max, total
	for _,t := range queryTool.queryTimes {
		// New min / max
		if t < minTime {
			minTime = t
		}
		if t > maxTime {
			maxTime = t
		}
		totalTime += t
		// fmt.Println("wwwwwwwwwwwwwww", t.String())
	}

	// Compute average
	avgTime := time.Duration(int64(totalTime) / int64(numQueries))

	// Sort & compute median
	var medianTime time.Duration
	sort.Sort(Duration(queryTool.queryTimes))
	if numQueries % 2 == 0 {
		medianTime = (queryTool.queryTimes[numQueries/2-1] + queryTool.queryTimes[numQueries/2]) / 2
	} else {
		medianTime = queryTool.queryTimes[numQueries/2]
	}

	// TODO: Remove this
	fmt.Println(queryTool.queryTimes)

	// Output
	fmt.Printf("\n%s\n", strings.Repeat("=",30))
	fmt.Printf("Queries run:  %d\n", numQueries)
	fmt.Printf(" Total time: %6.3fs\n", float64(totalTime)  / float64(time.Second))
	fmt.Printf("   Min time: %6.3fs\n", float64(minTime)    / float64(time.Second))
	fmt.Printf("   Max time: %6.3fs\n", float64(maxTime)    / float64(time.Second))
	fmt.Printf("   Avg time: %6.3fs\n", float64(avgTime)    / float64(time.Second))
	fmt.Printf("Median time: %6.3fs\n", float64(medianTime) / float64(time.Second))
	fmt.Println("")
}
