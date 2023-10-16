package migrations

import (
	"database/sql"
	"os"
	"path/filepath"
	"slices"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/supagrate/cli/internal/utils"
)

const MigrationsDirectory = utils.SupagrateDirectory + "/migrations"

type Migration struct {
	FileName string
	Up       string
	Down     string
}

func (migration Migration) Apply(db *sql.DB) {
	logrus.Info("Applying migration: " + migration.FileName)

	_, err := db.Exec(migration.Up)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	migration.Record(db)

	logrus.Info("Migration applied: " + migration.FileName)
}

func (migration Migration) Record(db *sql.DB) {
	logrus.Info("Recording migration: " + migration.FileName)

	_, err := db.Exec("insert into supagrate.migrations (name) values ($1)", migration.FileName)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	logrus.Info("Migration recorded: " + migration.FileName)
}

func (migration Migration) Rollback(db *sql.DB) {
	logrus.Info("Rolling back migration: " + migration.FileName)

	_, err := db.Exec(migration.Down)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	migration.Scratch(db)

	logrus.Info("Rollback complete: " + migration.FileName)
}

func (migration Migration) Scratch(db *sql.DB) {
	logrus.Info("Scratching migration: " + migration.FileName)

	_, err := db.Exec("delete from supagrate.migrations where name = $1", migration.FileName)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	logrus.Info("Migration scratched: " + migration.FileName)
}

type DBMigration struct {
	ID        string
	Name      string
	CreatedAt string
}

type Status struct {
	Name    string
	Present bool
}

func EnsureMigrationDirectoryExists(fs afero.Fs, migrationName string) string {
	migrationsDir := filepath.Join(".", MigrationsDirectory, utils.GetCurrentTimestamp()+"_"+migrationName)

	if err := utils.MkdirIfNotExist(fs, filepath.Dir(migrationsDir)); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	return migrationsDir
}

func ReadMigrationsFromDB(db *sql.DB) []DBMigration {
	rows, err := db.Query("select id, name, created_at from supagrate.migrations")
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	var dbMigrations []DBMigration

	for rows.Next() {
		var dbMigration DBMigration
		err := rows.Scan(&dbMigration.ID, &dbMigration.Name, &dbMigration.CreatedAt)

		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}

		dbMigrations = append(dbMigrations, dbMigration)
	}

	return dbMigrations
}

func ReadMigrationsFromFilesystem(fs afero.Fs) []Migration {
	migrations, err := afero.ReadDir(fs, MigrationsDirectory)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	var upMigrations []Migration

	for _, migration := range migrations {
		logrus.Info(migration.Name())

		if migration.IsDir() {
			upMigrations = append(upMigrations, readMigration(fs, migration.Name()))
		}
	}

	return upMigrations
}

func FindMissingMigrations(migrations []Migration, dbMigrations []DBMigration) []Migration {
	var missing []Migration

	for _, migration := range migrations {
		if index := slices.IndexFunc(dbMigrations, func(dbMigration DBMigration) bool { return dbMigration.Name == migration.FileName }); index == -1 {
			missing = append(missing, migration)
		}
	}

	return missing
}

func FindExistingMigrations(dbMigrations []DBMigration, migrations []Migration) []Migration {
	var missing []Migration

	for _, migration := range migrations {
		if index := slices.IndexFunc(dbMigrations, func(dbMigration DBMigration) bool { return dbMigration.Name == migration.FileName }); index > -1 {
			missing = append(missing, migration)
		}
	}

	return missing
}

func FindMigrationStatus(migrations []Migration, dbMigrations []DBMigration) []Status {
	var status []Status

	for _, migration := range migrations {
		if index := slices.IndexFunc(dbMigrations, func(dbMigration DBMigration) bool { return dbMigration.Name == migration.FileName }); index == -1 {
			status = append(status, Status{
				Name:    migration.FileName,
				Present: false,
			})
		} else {
			status = append(status, Status{
				Name:    migration.FileName,
				Present: true,
			})
		}
	}

	return status
}

func GenerateMigrationFile(fs afero.Fs, migrationDir string, filename string) {
	path := filepath.Join(migrationDir, filename)

	if err := utils.MkdirIfNotExist(fs, filepath.Dir(path)); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	f, err := fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	defer f.Close()
}

func readMigration(fs afero.Fs, migrationName string) Migration {
	upFilename := filepath.Join(MigrationsDirectory, migrationName, "up.sql")
	up, err := afero.ReadFile(fs, upFilename)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	downFilename := filepath.Join(MigrationsDirectory, migrationName, "down.sql")
	down, err := afero.ReadFile(fs, downFilename)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	return Migration{
		FileName: migrationName,
		Up:       string(up),
		Down:     string(down),
	}
}
