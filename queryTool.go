package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)
import _ "github.com/lib/pq"

const LINE_FORMAT_EXAMPLE string = "" +
	"Format:  host,startTimestamp,endTimestamp\n" +
	"Example: host_000000,2017-01-01 00:00:00,2017-01-01 01:00:00"

// Duration is sortable because it adheres to sort interface
type Duration []time.Duration
func (d Duration) Len() int       	  { return len(d) }
func (d Duration) Less(i, j int) bool { return d[i] < d[j] }
func (d Duration) Swap(i, j int)	  { d[i], d[j] = d[j], d[i] }

/*
   QueryTool is used to run benchmarked queries against the TimescaleDB instance.

   It uses an async worker multiqueue (with size of input concurrency) to run queries.
   Jobs are consistently hashed into one of the queues when they are enqueued.
*/
type QueryTool struct {
	db							*Database
	multiQueue				[]*Queue
	queryTimes				[]time.Duration
	mu		   				sync.Mutex	// Used for safe updating of queryTimes
	outputQueryResults	bool
}

// Returns an instance of QueryTool.
func NewQueryTool(db *Database, concurrency uint, outputQueryResults bool) *QueryTool {
	if concurrency <= 0 {
		fmt.Printf("Concurrency <= 0 (is %d), setting to 1\n", concurrency)
		concurrency = 1
	}

	// Create multi-queue
	queues := make([]*Queue, 0, concurrency)
	for i := 0; i < int(concurrency); i++ {
		queues = append(queues, NewQueue())
	}

	// Create QueryTool instance
	// Initial capacity of 256 to avoid some append() array copying using the provided query CSV.
	queryTool := QueryTool{
		db: db,
		multiQueue: queues,
		queryTimes: make([]time.Duration, 0, 256),
		mu: sync.Mutex{},
		outputQueryResults: outputQueryResults,
	}

	// Start queues
	queryTool.startMultiQueue()

	return &queryTool
}

// Returns the queue this key hashes into.
func (queryTool *QueryTool) getQueue(key string) *Queue {
	// Hash key (the hostname) & cast to int
	h := fnv.New32a()
	h.Write([]byte(key))
	hash := h.Sum32()
	// Get the queue # it will use
	bucket := int(hash) % len(queryTool.multiQueue)
	return queryTool.multiQueue[bucket]
}

// Starts all queues in the multiqueue.
func (queryTool *QueryTool) startMultiQueue() {
	// fmt.Println("Multiqueue starting!")
	for _,q := range queryTool.multiQueue {
		go q.Start()
	}
}

// Waits for all multiqueue wait groups to finish.
func (queryTool *QueryTool) waitAllMultiQueue() {
	// fmt.Println("Multiqueue waiting!")
	for _,q := range queryTool.multiQueue {
		q.Wait()
	}
}

// Stops all queues in the multiqueue.
func (queryTool *QueryTool) stopMultiQueue() {
	// fmt.Println("Multiqueue stopping!")
	for _,q := range queryTool.multiQueue {
		q.Stop()
	}
}

// Runs the tool using CSV file as input, prints runtime stats when finished.
// The file is read line-by-line, each line being translated into a db query,
// & immediately submitted to the multiqueue for execution.
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
		parts, err := queryTool.parseLine(line)
		if err != nil {
			fmt.Println(err)
			continue
		}
		host, start, end := parts[0], parts[1], parts[2]

		// Add job to multiqueue
		queue := queryTool.getQueue(host)
		queue.Enqueue(Job{start, end, host, queryTool.runQuery, jobNum})
		jobNum += 1
	}

	// Wait on all multiqueues, then stop them
	queryTool.waitAllMultiQueue()
	queryTool.stopMultiQueue()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning file:", err)
	}

	queryTool.printQueryTimeStats()
}

// Runs the tool in REPL-like mode where the user manually enters each line.
// Every line the user submits is translated into a db query, & immediately
// submitted to the multiqueue for execution.
func (queryTool *QueryTool) RunWithManualInput(filePath string) {
	// Print instructions & example line format.
	fmt.Println("You're running in interactive mode. Enter lines in the format:")
	fmt.Println(LINE_FORMAT_EXAMPLE)
	fmt.Println(`Type "exit" to exit`)

	scanner := bufio.NewScanner(os.Stdin)

	jobNum := 0
	for {
		fmt.Print("> ")

		scanned := scanner.Scan()
		if !scanned {
			panic(scanner.Err())
		}

		line := scanner.Text()

		if line == "exit" {
			break
		} else if line == "" {
			continue
		} else {
			parts, err := queryTool.parseLine(line)
			if err != nil {
				fmt.Println(err)
				continue
			}
			host, start, end := parts[0], parts[1], parts[2]

			// Add job to multiqueue
			queue := queryTool.getQueue(host)
			queue.Enqueue(Job{start, end, host, queryTool.runQuery, jobNum})
			jobNum += 1
		}
	}

	// Wait on all multiqueues, then stop them
	queryTool.waitAllMultiQueue()
	queryTool.stopMultiQueue()

	queryTool.printQueryTimeStats()
}

// Parses & returns the line, prints helpful error if invalid
func (queryTool *QueryTool) parseLine(line string) ([]string, error) {
	parts := strings.Split(line, ",")

	// Validate length
	if len(parts) != 3 {
		msg := fmt.Sprintf("Invalid line: %s", line)
		return []string{}, errors.New(msg)
	}

	// Validate timestamp fields
	start, end := parts[1], parts[2]
	if !queryTool.isValidTimestamp(start) {
		msg := fmt.Sprintf("Invalid timestamp: %s", start)
		return []string{}, errors.New(msg)
	}
	if !queryTool.isValidTimestamp(end) {
		msg := fmt.Sprintf("Invalid timestamp: %s", end)
		return []string{}, errors.New(msg)
	}

	return parts, nil
}

// Validates timestamp is of format yyyy-mm-dd
func (queryTool *QueryTool) isValidTimestamp(ts string) bool {
	format := "2006-01-02 15:04:05"
	_, err := time.Parse(format, ts)
	return err == nil
}

// Runs the given query in the db, prints the results, & returns the runtime of the query operation.
func (queryTool *QueryTool) runQuery(job Job) time.Duration {
	// Setup
	query := readFile("./sql/query_cpuMinMaxByMin.sql")
	conn := queryTool.db.GetConnection()
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
	defer rows.Close()
	queryEnd := time.Now()

	if err != nil {
		panic(err)
	}

	// Optionally print query results
	if queryTool.outputQueryResults {
		queryTool.printQueryResults(rows)
	}

	// Calculate runtime & return it
	elapsedTime := queryEnd.Sub(queryStart)

	// Use mutex to lock the shared resource to avoid race conditions
	// NOTE: It seemed to work 100% of the time without this, but the -race flag was correctly letting me know that
	// race conditions were happening. The overhead on this seems negligable in my tests. Better safe than sorry. Don't
	// make anybody have to debug race conditions!
	queryTool.mu.Lock()
	queryTool.queryTimes = append(queryTool.queryTimes, elapsedTime)
	queryTool.mu.Unlock()

	return elapsedTime
}

// Loops over query result rows & prints each one.
// NOTE: Does NOT close rows! The caller is responsible for that.
func (queryTool *QueryTool) printQueryResults(rows *sql.Rows) {
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

		fmt.Println(host, ts, cpuMin, cpuMax)
		count += 1
	}
	fmt.Println("")
}

// Prints time stats related to how long the queries took
func (queryTool *QueryTool) printQueryTimeStats() {
	// Setup
	numQueries := len(queryTool.queryTimes)
	var minTime time.Duration
	var maxTime time.Duration
	// Set min & max to 0 if there's no times present
	if numQueries == 0 {
		minTime = 0 * time.Second
		maxTime = 0 * time.Second
	} else {
		// Prime min & max as 0th element
		minTime = queryTool.queryTimes[0]
		maxTime = queryTool.queryTimes[0]
	}

	// Compute min, max, total
	var totalTime time.Duration
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
	var avgTime time.Duration
	if numQueries == 0 {
		avgTime = 0 * time.Second
	} else {
		avgTime = time.Duration(int64(totalTime) / int64(numQueries))
	}

	// Sort & compute median
	var medianTime time.Duration
	sort.Sort(Duration(queryTool.queryTimes))
	if numQueries == 0 {
		medianTime = 0 * time.Second
	} else {
		if numQueries % 2 == 0 {
			medianTime = (queryTool.queryTimes[numQueries/2-1] + queryTool.queryTimes[numQueries/2]) / 2
		} else {
			medianTime = queryTool.queryTimes[numQueries/2]
		}
	}

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
	fmt.Println("Note: That is *not* tool run time, it's total query time! If queries ran in parallel,\nthe value may be larger than expected.")
	fmt.Println("")
}

// Returns file contents as a string
func readFile(filePath string) string {
	query, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	queryStr := string(query)
	return queryStr
}
