package diff

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	_init "github.com/supagrate/cli/internal/init"
	"github.com/supagrate/cli/internal/utils"
	"github.com/supagrate/cli/migrations"
)

const migraImage = "ghcr.io/supabase/migra:3.0.1663481299"

// ErrorHook is a logrus hook that ensures error messages are always printed
type ErrorHook struct {
	originalOutput *os.File
}

// Levels returns the levels this hook should be fired on
func (h *ErrorHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

// Fire is called when a log event occurs
func (h *ErrorHook) Fire(entry *logrus.Entry) error {
	line := fmt.Sprintf("%s%s\n", utils.Red("Error: "), entry.Message)
	h.originalOutput.WriteString(line)
	return nil
}

// convertToURL converts a PostgreSQL connection string from key=value format to URL format
func convertToURL(connStr string) string {
	// If it's already a URL, return as is but ensure it uses postgresql://
	if strings.HasPrefix(connStr, "postgres://") {
		return strings.Replace(connStr, "postgres://", "postgresql://", 1)
	}
	if strings.HasPrefix(connStr, "postgresql://") {
		return connStr
	}

	// Parse key=value pairs
	parts := strings.Fields(connStr)
	params := make(map[string]string)
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			params[kv[0]] = kv[1]
		}
	}

	// Construct URL
	url := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		params["user"],
		params["password"],
		params["host"],
		params["port"],
		params["dbname"])

	// Add SSL mode if present
	if sslmode, ok := params["sslmode"]; ok {
		url += "?sslmode=" + sslmode
	}

	return url
}

func Run(cmd *cobra.Command, fs afero.Fs) {
	_init.EnsureInitialization()

	// Get debug flag
	debug, _ := cmd.Flags().GetBool("debug")

	// Setup error hook to ensure errors are always visible
	errorHook := &ErrorHook{originalOutput: os.Stderr}
	logrus.AddHook(errorHook)

	if !debug {
		logrus.SetOutput(io.Discard) // Only suppress non-error logs
	}

	// Get schemas to compare
	schemas, _ := cmd.Flags().GetStringSlice("schemas")
	if len(schemas) == 0 {
		schemas = []string{"public"}
	}

	// Create migrations directory if it doesn't exist
	if err := utils.MkdirIfNotExist(fs, migrations.MigrationsDirectory); err != nil {
		fmt.Println(utils.Red("Error:"), err)
		os.Exit(1)
	}

	// Get the migration name from the -f flag
	migrationName := cmd.Flag("file").Value.String()

	// Connect to the main database
	connectionString := cmd.Flag("connection").Value.String()
	var conn *utils.Connection
	if connectionString != "" {
		conn = &utils.Connection{Connection: utils.ParseConnectionString(connectionString)}
	} else {
		conn = &utils.Connection{Connection: nil}
	}
	db := utils.ConnectDatabase(*conn)
	defer db.Close()

	// Debug: Check tables in main database
	var tables []string
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		fmt.Println(utils.Red("Error:"), "failed to check tables in main database:", err)
		os.Exit(1)
	}
	defer rows.Close()
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			fmt.Println(utils.Red("Error:"), "failed to scan table name:", err)
			os.Exit(1)
		}
		tables = append(tables, tableName)
	}
	logrus.Debugf("Tables in main database: %v", tables)

	// Debug: Check posts table structure if it exists
	if slices.Contains(tables, "posts") {
		var tableInfo string
		err := db.QueryRow(`
			SELECT pg_catalog.pg_get_tabledef('public.posts'::regclass::oid)
		`).Scan(&tableInfo)
		if err != nil {
			logrus.Debugf("Failed to get posts table info: %v", err)
		} else {
			logrus.Debugf("Posts table in main DB:\n%s", tableInfo)
		}
	}

	// Create a shadow database for comparison
	shadowDB, err := createShadowDatabase(db)
	if err != nil {
		fmt.Println(utils.Red("Error:"), err)
		os.Exit(1)
	}
	defer dropShadowDatabase(db, shadowDB)

	// Apply migrations to shadow database
	if err := applyMigrationsToShadowDB(fs, shadowDB); err != nil {
		fmt.Println(utils.Red("Error:"), err)
		os.Exit(1)
	}

	// Compare schemas and get the diff
	mainDBURL := convertToURL(utils.GetDatabaseURL(*conn).DSN)
	diff, err := compareSchemas(mainDBURL, shadowDB, schemas, debug)
	if err != nil {
		fmt.Println(utils.Red("Error:"), err)
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println(utils.Yellow("Database is up to date"))
		return
	}

	// If migration name is provided, create migration files
	if migrationName != "" {
		createMigrationFiles(fs, migrationName, diff)
	} else {
		// Just print the diff
		fmt.Println(utils.Green("Schema differences detected:"))
		fmt.Println(diff)
	}

	// Restore logging
	if !debug {
		logrus.SetOutput(os.Stderr)
	}
}

func createShadowDatabase(db *sql.DB) (string, error) {
	timestamp := time.Now().UTC().UnixNano()
	shadowDBName := fmt.Sprintf("shadow_db_%d", timestamp)

	// First ensure the shadow database doesn't exist
	_, err := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", shadowDBName))
	if err != nil {
		return "", fmt.Errorf("failed to drop existing shadow database: %v", err)
	}

	// Create a fresh shadow database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", shadowDBName))
	if err != nil {
		return "", fmt.Errorf("failed to create shadow database: %v", err)
	}

	return shadowDBName, nil
}

func dropShadowDatabase(db *sql.DB, shadowDBName string) {
	// Terminate all connections to the shadow database first
	_, err := db.Exec(fmt.Sprintf(`
		SELECT pg_terminate_backend(pid) 
		FROM pg_stat_activity 
		WHERE datname = '%s' 
		AND pid <> pg_backend_pid()`, shadowDBName))
	if err != nil {
		logrus.Warnf("Failed to terminate connections to shadow database: %v", err)
	}

	// Drop the database
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", shadowDBName))
	if err != nil {
		logrus.Warnf("Failed to drop shadow database: %v", err)
	}
}

func applyMigrationsToShadowDB(fs afero.Fs, shadowDBName string) error {
	// Get all migrations from filesystem
	fsMigrations := migrations.ReadMigrationsFromFilesystem(fs)
	logrus.Infof("Found %d migrations in filesystem", len(fsMigrations))
	for _, m := range fsMigrations {
		logrus.Debugf("Migration %s content:\n%s", m.FileName, m.Up)
	}

	// Connect to shadow database
	shadowURL := fmt.Sprintf("postgresql://postgres:postgres@localhost:5432/%s?sslmode=disable", shadowDBName)
	shadowConn := utils.Connection{
		Connection: utils.ParseConnectionString(shadowURL),
	}
	shadowDB := utils.ConnectDatabase(shadowConn)
	defer shadowDB.Close()

	// Create supagrate schema and migrations table
	logrus.Info("Creating supagrate schema and migrations table")
	utils.EnsureMigrationTable(shadowDB)

	// Apply each migration
	for _, migration := range fsMigrations {
		logrus.Infof("Applying migration %s to shadow database", migration.FileName)
		if err := migration.Apply(shadowDB); err != nil {
			return fmt.Errorf("failed to apply migration %s to shadow database: %v", migration.FileName, err)
		}
		logrus.Infof("Successfully applied migration %s", migration.FileName)
	}

	// Debug: Check tables in shadow database
	var tables []string
	rows, err := shadowDB.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		return fmt.Errorf("failed to check tables in shadow database: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %v", err)
		}
		tables = append(tables, tableName)
	}
	logrus.Debugf("Tables in shadow database: %v", tables)

	// Debug: Check posts table structure if it exists
	if slices.Contains(tables, "posts") {
		var tableInfo string
		err := shadowDB.QueryRow(`
			SELECT pg_catalog.pg_get_tabledef('public.posts'::regclass::oid)
		`).Scan(&tableInfo)
		if err != nil {
			logrus.Debugf("Failed to get posts table info: %v", err)
		} else {
			logrus.Debugf("Posts table in shadow DB:\n%s", tableInfo)
		}
	}

	return nil
}

func compareSchemas(mainDBURL string, shadowDBName string, schemas []string, debug bool) (string, error) {
	shadowDBURL := fmt.Sprintf("postgresql://postgres:postgres@localhost:5432/%s?sslmode=disable", shadowDBName)

	// Initialize Docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer cli.Close()

	// Pull migra image if not exists
	reader, err := cli.ImagePull(ctx, migraImage, types.ImagePullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull migra image: %v", err)
	}
	defer reader.Close()

	if debug {
		io.Copy(os.Stdout, reader) // Show pull progress only in debug mode
	} else {
		io.Copy(io.Discard, reader) // Discard output in normal mode
	}

	// Build migra command with schema arguments
	migraCmd := []string{"migra", "--unsafe"}
	for _, schema := range schemas {
		migraCmd = append(migraCmd, "--schema", schema)
	}
	// Compare shadow DB (source) to main DB (target) to see what changes need to be applied
	migraCmd = append(migraCmd, shadowDBURL, mainDBURL)

	// Create container config
	config := &container.Config{
		Image: migraImage,
		Cmd:   migraCmd,
	}

	// Create host config with network mode host to access local postgres
	hostConfig := &container.HostConfig{
		NetworkMode: "host",
	}

	// Create container
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	// Start container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}

	// Wait for container to finish
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", fmt.Errorf("error waiting for container: %v", err)
		}
	case <-statusCh:
	}

	// Get container logs
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %v", err)
	}
	defer out.Close()

	// Read and demultiplex the logs
	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, out)
	if err != nil {
		return "", fmt.Errorf("failed to read container logs: %v", err)
	}

	// Cleanup container
	err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		logrus.Warnf("Failed to remove container: %v", err)
	}

	// Check for errors in stderr
	if stderr.Len() > 0 {
		stderrStr := stderr.String()
		// Always print the error regardless of debug mode
		fmt.Println(utils.Red("\nPostgreSQL Error:"))
		fmt.Println(stderrStr)
		return "", fmt.Errorf("database error occurred")
	}

	// Clean up the output and check for Postgres errors in stdout
	output := stdout.String()
	output = strings.TrimSpace(output)

	// Check for Postgres errors in the output
	if strings.Contains(strings.ToLower(output), "error") ||
		strings.Contains(strings.ToLower(output), "fatal") {
		fmt.Println(utils.Red("\nPostgreSQL Error:"))
		fmt.Println(output)
		return "", fmt.Errorf("database error occurred")
	}

	if output == "" {
		return "", nil
	}

	return output + "\n", nil
}

func createMigrationFiles(fs afero.Fs, migrationName string, diff string) {
	// Create migration directory with timestamp using the same UTC function as 'migrate new'
	timestamp := utils.GetCurrentTimestamp()
	migrationDir := filepath.Join(migrations.MigrationsDirectory, fmt.Sprintf("%s_%s", timestamp, migrationName))

	// Ensure the migrations directory exists
	if err := utils.MkdirIfNotExist(fs, filepath.Dir(migrationDir)); err != nil {
		logrus.Fatalf("Failed to create migrations directory: %v", err)
	}

	// Create the migration directory
	if err := utils.MkdirIfNotExist(fs, migrationDir); err != nil {
		logrus.Fatalf("Failed to create migration directory: %v", err)
	}

	// Create up.sql with the diff
	upPath := filepath.Join(migrationDir, "up.sql")
	err := afero.WriteFile(fs, upPath, []byte(diff), 0644)
	if err != nil {
		logrus.Fatalf("Failed to create up.sql: %v", err)
	}

	// Create empty down.sql (user will need to fill this manually)
	downPath := filepath.Join(migrationDir, "down.sql")
	err = afero.WriteFile(fs, downPath, []byte("-- Add rollback SQL here\n"), 0644)
	if err != nil {
		logrus.Fatalf("Failed to create down.sql: %v", err)
	}

	fmt.Println(utils.Green("New migration created at ") + migrationDir)
}
