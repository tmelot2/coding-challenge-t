package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// Configs used to connect to the database
type DatabaseConfig struct {
	username string
	password string
	url      string // Includes port #
	dbName   string
	options  string
}

// Reads & returns database config from the .env file
func getDatabaseConfig(envFilePath string) *DatabaseConfig {
	// Open the file
	file, err := os.Open(envFilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Setup
	var username string
	var password string
	var url string
	var dbName string
	var options string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore blank lines
		if line == "" {
			continue
		}

		// Split line, format is var=val
		split := strings.SplitN(line, "=", 2)
		// Ignore invalid lines
		if len(split) != 2 {
			continue
		}

		// Get values
		switch split[0] {
		case "DB_USER":
			username = split[1]
		case "DB_PASS":
			password = split[1]
		case "DB_URL":
			url = split[1]
		case "DB_DATABASE":
			dbName = split[1]
		case "DB_OPTIONS":
			options = split[1]
		}
	}

	return &DatabaseConfig{
		username: username,
		password: password,
		url:      url,
		dbName:   dbName,
		options:  options,
	}
}

// Database is used to get connections to the db
type Database struct {
	config	*DatabaseConfig
}

// Returns a new Database instance with the config loaded from the given file path
func NewDatabase(envFilePath string) *Database {
	dbConfig := getDatabaseConfig(envFilePath)
	return &Database{dbConfig}
}

// Returns a database connection
func (db *Database) GetConnection() *sql.DB {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?%s",
		db.config.username,
		db.config.password,
		db.config.url,
		db.config.dbName,
		db.config.options,
	)
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
