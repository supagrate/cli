package utils

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xo/dburl"

	_ "github.com/lib/pq"
)

// Local Supabase: postgresql://postgres:postgres@localhost:54322/postgres
type Connection struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type ConnectionEnv struct {
	Env  string
	Flag string
}

func (c Connection) ConnectionString() string {
	connectionString := "postgresql://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.Name

	if c.Host == "localhost" {
		connectionString += "?sslmode=disable"
	}

	return connectionString
}

func (c Connection) URL() *dburl.URL {
	return ParseConnectionString(c.ConnectionString())
}

func UseDBFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("db-host", "d", "localhost", "Database host")
	cmd.PersistentFlags().IntP("db-port", "o", 54322, "Database port")
	cmd.PersistentFlags().StringP("db-user", "u", "postgres", "Database user")
	cmd.PersistentFlags().StringP("db-password", "p", "postgres", "Database password")
	cmd.PersistentFlags().StringP("db-name", "n", "postgres", "Database name")
}

func ParseConnectionString(connectionString string) *dburl.URL {
	u, err := dburl.Parse(connectionString)

	if err != nil {
		logrus.Fatal("Could not parse this connection string")
	}

	return u
}

func ConnectDatabase(c Connection) *sql.DB {
	connection := c.URL().DSN

	db, err := sql.Open(c.URL().Driver, connection)

	if err != nil {
		logrus.Fatal(err)
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

func UseDBEnvironmentVariables(cmd *cobra.Command) {
	env := []ConnectionEnv{
		{"DB_HOST", "db-host"},
		{"DB_PORT", "db-port"},
		{"DB_USER", "db-user"},
		{"DB_PASSWORD", "db-password"},
		{"DB_NAME", "db-name"},
	}

	for _, e := range env {
		if value := os.Getenv(e.Env); value != "" {
			logrus.Infoln("Using " + e.Env + " environment variable")
			cmd.Flags().Set(e.Flag, value)
		}
	}
}
