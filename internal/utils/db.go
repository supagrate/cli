package utils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/xo/dburl"

	_ "github.com/lib/pq"
)

func init() {
	// Try to load .env file from different possible locations
	envPaths := []string{
		".env",
		"../.env",
		filepath.Join(os.Getenv("HOME"), ".supagrate", ".env"),
	}

	loaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			logrus.Debugf("Loaded environment from %s", path)
			loaded = true
			break
		}
	}

	if !loaded {
		logrus.Debug("No .env file found in search paths")
	}
}

type Connection struct {
	Connection *dburl.URL
}

func addSSLModeIfLocalhost(connStr string) string {
	logrus.Debugf("Processing connection string: %s", connStr)
	if (containsLocalhost(connStr) || contains127(connStr)) && !containsSSLMode(connStr) {
		logrus.Debugf("Local connection detected, adding sslmode=disable")
		if hasQuery(connStr) {
			return connStr + "&sslmode=disable"
		}
		return connStr + "?sslmode=disable"
	}
	return connStr
}

func containsLocalhost(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "localhost") ||
		strings.Contains(s, "host=localhost") ||
		strings.Contains(s, "[::1]") ||
		strings.Contains(s, "127.0.0.1")
}

func contains127(s string) bool {
	return strings.Contains(strings.ToLower(s), "127.0.0.1")
}

func containsSSLMode(s string) bool {
	return strings.Contains(strings.ToLower(s), "sslmode=")
}

func hasQuery(s string) bool {
	return strings.Contains(s, "?")
}

func ParseConnectionString(connectionString string) *dburl.URL {
	connectionString = addSSLModeIfLocalhost(connectionString)
	u, err := dburl.Parse(connectionString)

	if err != nil {
		logrus.Fatal("Could not parse this connection string")
	}

	return u
}

// GetDatabaseURL returns a *dburl.URL using DATABASE_URL if set, otherwise uses the Connection field.
func GetDatabaseURL(c Connection) *dburl.URL {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		logrus.Infoln("Using DATABASE_URL environment variable")
		logrus.Debugf("Original DATABASE_URL: %s", dbURL)
		parsed := ParseConnectionString(dbURL)
		logrus.Debugf("Final DSN: %s", parsed.DSN)
		return parsed
	}
	if c.Connection != nil {
		return c.Connection
	}
	logrus.Fatal("No database connection string provided")
	return nil
}

func ConnectDatabase(c Connection) *sql.DB {
	url := GetDatabaseURL(c)
	connection := url.DSN
	logrus.Debugf("Connecting with DSN: %s", connection)

	db, err := sql.Open(url.Driver, connection)

	if err != nil {
		logrus.Fatal(err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

func EnsureSchema(db *sql.DB) {
	result, err := db.Query("create schema if not exists supagrate")

	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	defer result.Close()
}

func EnsureMigrationTable(db *sql.DB) {
	EnsureSchema(db)

	result, err := db.Query("create table if not exists supagrate.migrations (id uuid primary key default gen_random_uuid(), name VARCHAR(255) not null, created_at TIMESTAMP not null default current_timestamp)")

	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	defer result.Close()
}

func ResetMigrationTable(db *sql.DB) {
	result, err := db.Query("drop table if exists supagrate.migrations")

	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	defer result.Close()
}

func ResetPublicSchema(db *sql.DB) {
	logrus.Info("Resetting public schema...")

	_, err := db.Exec("drop schema public cascade; create schema public;")

	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
}

func Test(db *sql.DB) {
	rows, err := db.Query("select * from test")

	if err != nil {
		logrus.Panic(err)
	}

	for rows.Next() {
		var id int
		var created_at string
		err = rows.Scan(&id, &created_at)
		fmt.Println(id, created_at)
	}

	defer rows.Close()
}
