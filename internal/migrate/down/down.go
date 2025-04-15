package down

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	_init "github.com/supagrate/cli/internal/init"
	"github.com/supagrate/cli/internal/utils"
	"github.com/supagrate/cli/migrations"
)

func Run(cmd *cobra.Command, fs afero.Fs) {
	_init.EnsureInitialization()

	count := utils.ParseCount(cmd.Flag("count").Value.String())

	connectionString := cmd.Flag("connection").Value.String()
	var conn *utils.Connection
	if connectionString != "" {
		conn = &utils.Connection{Connection: utils.ParseConnectionString(connectionString)}
	} else {
		conn = &utils.Connection{Connection: nil}
	}
	db := utils.ConnectDatabase(*conn)

	utils.EnsureMigrationTable(db)

	// Fetch all existing migrations in DB
	dbMigrations := migrations.ReadMigrationsFromDB(db)

	// Fetch all existing migrations in filesystem
	fsMigrations := migrations.ReadMigrationsFromFilesystem(fs)

	existing := migrations.FindExistingMigrations(dbMigrations, fsMigrations)

	// If count higher than missing migrations, set count to missing migrations
	if count <= -1 || count > len(existing) {
		count = len(existing)
	}

	slices.Reverse(existing)
	existing = existing[:count]

	if len(existing) == 0 {
		fmt.Println(utils.Red("No migrations to rollback"))
		os.Exit(0)
	}

	// Rollback the migrations from the DB (max to the count)
	for _, migration := range existing {
		fmt.Println("Rolling back " + utils.Yellow(migration.FileName) + "...")
		migration.Rollback(db)
	}

	fmt.Println(utils.Green("Successfully rolled back " + fmt.Sprintf("%d", count) + " migrations"))

	db.Close()
}
