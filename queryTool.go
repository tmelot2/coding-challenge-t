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

// TODO: Doc
type QueryTool struct {
	queryTimes []time.Duration
}

// TODO: Doc
func newQueryTool() *QueryTool {
	queryTool := QueryTool{make([]time.Duration, 0, 256)}
	return &queryTool
}

// TODO: Doc
func (queryTool *QueryTool) RunWithCsvFile(filePath string) {
	// Open CSV file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening CSV file:", err)
		return
	}
	defer file.Close()

	// Cache query file
	query := readFile("query_cpuMinMaxByMin.sql")

	// Read & process CSV line-by-line
	scanner := bufio.NewScanner(file)
	scanner.Scan() // Ignore 1st line (it's a header)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		host, start, end := parts[0], parts[1], parts[2]
		queryTime := queryTool.runQuery(query, start, end, host)
		fmt.Println(queryTime)
		queryTool.queryTimes = append(queryTool.queryTimes, queryTime)

		fmt.Println("\n==============================\n")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning file:", err)
	}

	queryTool.printQueryTimeStats()
}

// TODO: Doc
func (queryTool *QueryTool) runQuery(query, start, end, host string) time.Duration {
	conn := queryTool.getDatabaseConnection()
	defer conn.Close()

	queryStart := time.Now()
	rows, err := conn.Query(query, start, end, host)
	queryEnd := time.Now()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	fmt.Println(host)

	// ts (ts), host (str), usage (double)
	for rows.Next() {
		var host string
		var ts time.Time
		var cpuMin float64
		var cpuMax float64

		err := rows.Scan(&host, &ts, &cpuMin, &cpuMax)
		if err != nil {
			panic(err)
		}

		// fmt.Println(ts, cpuMin, cpuMax)
	}

	elapsedTime := queryEnd.Sub(queryStart)
	return elapsedTime
}

// TODO: Refactor to ConnectionManager or something, & pass in a ref of that to QueryTool constructor
func (queryTool *QueryTool) getDatabaseConnection() *sql.DB {
	// TODO: Pull from .env file
	username := "tsdbadmin"
	password := "oda9b95cubho3tqq"
	host := "aqadz0sy32.tzug8uusr7.tsdb.cloud.timescale.com:39894"
	db := "tsdb"
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

// TODO: Doc
func (queryTool *QueryTool) printQueryTimeStats() {
	numQueries := len(queryTool.queryTimes)
	// Prime min & max as 0th element
	minTime := queryTool.queryTimes[0]
	maxTime := queryTool.queryTimes[0]
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
		fmt.Println(t.String())
	}

	// Calculate average & median
	avgTime := time.Duration(int64(totalTime) / int64(numQueries))

	// Sort & get median
	var medianTime time.Duration
	sort.Sort(Duration(queryTool.queryTimes))
	if numQueries % 2 == 0 {
		medianTime = (queryTool.queryTimes[numQueries/2-1] + queryTool.queryTimes[numQueries/2]) / 2
	} else {
		medianTime = queryTool.queryTimes[numQueries/2]
	}

	fmt.Println(queryTool.queryTimes)

	fmt.Println("\n%s\n", strings.Repeat("=",30))
	fmt.Printf("Queries run: %d\n", numQueries)
	fmt.Printf(" Total time: %6.3fs\n", float64(totalTime)  / float64(time.Second))
	fmt.Printf("   Min time: %6.3fs\n", float64(minTime)   / float64(time.Second))
	fmt.Printf("   Max time: %6.3fs\n", float64(maxTime)   / float64(time.Second))
	fmt.Printf("   Avg time: %6.3fs\n", float64(avgTime)    / float64(time.Second))
	fmt.Printf("Median time: %6.3fs\n", float64(medianTime) / float64(time.Second))
	fmt.Println("")
}

// Returns the bucket index this key hashes into for the given number of buckets
func (queryTool *QueryTool) getBucketIndex(key string, numBuckets int) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	hash := h.Sum32()
	bucket := int(hash) % numBuckets
	return bucket
}