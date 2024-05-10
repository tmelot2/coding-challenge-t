package main

// TODO: Rename this file

import (
	"bufio"
	"database/sql"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

import _ "github.com/lib/pq"

func loadCsvFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // Ignore 1st line (it's a header)
	buckets := make(map[int]int)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		host, start, end := parts[0], parts[1], parts[2]
		fmt.Printf("%s (%s - %s)\n", host, start, end)

		bucket := getBucketIndex(host, 4)
		buckets[bucket] += 1
	}

	for k,v := range buckets {
		fmt.Printf("Bucket %d count: %d\n", k, v)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning file:", err)
	}
}

// Returns the bucket index this key hashes into for the given number of buckets
func getBucketIndex(key string, numBuckets int) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	hash := h.Sum32()
	bucket := int(hash) % numBuckets
	return bucket
}

func connectToDatabase() *sql.DB {
	// TODO: Pull from .env file
	username := "username"
	password := "password"
	host := "host"
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

func loadCpuMinMaxQuery() string {
	query, err := ioutil.ReadFile("query_cpuMinMaxByMin.sql")
	if err != nil {
		panic(err)
	}
	queryStr := string(query)
	return queryStr
}

func main() {
	loadCsvFile("query_params.csv")

	host := "host_000014"
	start := "2017-01-02 22:55:00"
	end := "2017-01-02 22:59:00"

	db := connectToDatabase()
	defer db.Close()

	query := loadCpuMinMaxQuery()
	rows, err := db.Query(query, start, end, host)
	if err != nil {
		panic(err)
	}
	defer rows.Close()


	// ts (ts), host (str), usage (double)
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
	fmt.Println("Count:", count)
}
