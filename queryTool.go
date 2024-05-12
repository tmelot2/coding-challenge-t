package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"strings"
	"sync"
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
	mu		    sync.Mutex
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
	queryTool := QueryTool{queues, make([]time.Duration, 0, 256), sync.Mutex{}}

	// Start queues
	queryTool.startMultiQueue()

	return &queryTool
}

// Returns the queue this key hashes into.
func (queryTool *QueryTool) getQueue(key string) *Queue {
	h := fnv.New32a()
	h.Write([]byte(key))
	hash := h.Sum32()
	bucket := int(hash) % len(queryTool.multiQueue)
	return queryTool.multiQueue[bucket]
}

// Starts all queues in the multiqueue.
func (queryTool *QueryTool) startMultiQueue() {
	fmt.Println("Multiqueue starting!")
	for _,q := range queryTool.multiQueue {
		go q.Start()
	}
}

// Waits for all multiqueue wait groups to finish.
func (queryTool *QueryTool) waitAllMultiQueue() {
	fmt.Println("Multiqueue waiting!")
	for _,q := range queryTool.multiQueue {
		q.Wait()
	}
}

// Stops all queues in the multiqueue.
func (queryTool *QueryTool) stopMultiQueue() {
	fmt.Println("Multiqueue stopping!")
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

	jobNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		host, start, end := parts[0], parts[1], parts[2]

		queue := queryTool.getQueue(host)
		queue.Enqueue(Job{start, end, host, queryTool.runQuery, jobNum})
		jobNum += 1
	}

	// Wait on all multiqueues
	queryTool.waitAllMultiQueue()

	// Stop all queues in the multiqueue
	queryTool.stopMultiQueue()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning file:", err)
	}

	queryTool.printQueryTimeStats()
}

// Runs the given query in the db, prints the results, & returns the runtime of the query operation.
func (queryTool *QueryTool) runQuery(job Job) time.Duration {
	// Setup
	query := readFile("query_cpuMinMaxByMin.sql")
	conn := queryTool.getDatabaseConnection()
	defer conn.Close()

	// Prepare query
	stmt, err := conn.Prepare(query)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// Run the query & time how long it takes
	queryStart := time.Now()
	rows, err := stmt.Query(job.start, job.end, job.host)
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

	// Calculate runtime & return it
	elapsedTime := queryEnd.Sub(queryStart)

	// Use mutex to lock the shared resource to avoid race conditions
	// NOTE: It seemed to work 100% of the time without this, but the -race flag was correctly letting me know that
	// race conditions were happening. The overhead on this seems negligable in my tests. Better safe than sorry!
	queryTool.mu.Lock()
	queryTool.queryTimes = append(queryTool.queryTimes, elapsedTime)
	queryTool.mu.Unlock()

	return elapsedTime
}

// Connects to the db & returns the connection
// TODO: Refactor to ConnectionManager or something, & pass in a ref of that to QueryTool constructor
func (queryTool *QueryTool) getDatabaseConnection() *sql.DB {
	// TODO: Pull from .env file
	// postgres://tsdbadmin@aqadz0sy32.tzug8uusr7.tsdb.cloud.timescale.com:31633/tsdb?sslmode=require
	username := "username"
	password := "password"
	host := "host"
	db := "tsdb_transaction"
	args := "sslmode=require"

	connectionString := fmt.Sprintf("postgres://%s:%s@%s/%s?%s", username, password, host, db, args)
	conn, err := sql.Open("postgres", connectionString)
    conn.SetMaxOpenConns(10)
    conn.SetMaxIdleConns(10)
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
	}

	// Compute average
	avgTime := time.Duration(int64(totalTime) / int64(numQueries))

	// TODO: Remove this
	fmt.Println("Pre-sorted query times:", queryTool.queryTimes)

	// Sort & compute median
	var medianTime time.Duration
	sort.Sort(Duration(queryTool.queryTimes))
	if numQueries % 2 == 0 {
		medianTime = (queryTool.queryTimes[numQueries/2-1] + queryTool.queryTimes[numQueries/2]) / 2
	} else {
		medianTime = queryTool.queryTimes[numQueries/2]
	}
	fmt.Println("Sorted query times:", queryTool.queryTimes)

	// TODO: Remove this
	// fmt.Println(queryTool.queryTimes)

	// Output
	fmt.Printf("\n%s\n", strings.Repeat("=",30))
	fmt.Printf("Concurrency:  %d\n", len(queryTool.multiQueue))
	fmt.Printf("Queries run:  %d\n", numQueries)
	fmt.Printf(" Total time: %6.3fs\n", float64(totalTime)  / float64(time.Second))
	fmt.Printf("   Min time: %6.3fs\n", float64(minTime)    / float64(time.Second))
	fmt.Printf("   Max time: %6.3fs\n", float64(maxTime)    / float64(time.Second))
	fmt.Printf("   Avg time: %6.3fs\n", float64(avgTime)    / float64(time.Second))
	fmt.Printf("Median time: %6.3fs\n", float64(medianTime) / float64(time.Second))
	fmt.Println("")
}
