package up

import (
	"fmt"
	"os"

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

	// Compare the two and find the missing migrations
	missing := migrations.FindMissingMigrations(fsMigrations, dbMigrations)

	// If count higher than missing migrations, set count to missing migrations
	if count <= -1 || count > len(missing) {
		count = len(missing)
	}

	// Limit the number of migrations to run
	missing = missing[:count]

	if len(missing) == 0 {
		fmt.Println(utils.Red("No migrations to apply"))
		os.Exit(0)
	}

	// Run the migrations that are missing from the DB (max to the count)
	for _, migration := range missing {
		fmt.Println("Applying " + utils.Yellow(migration.FileName) + "...")
		migration.Apply(db)
	}

	fmt.Println(utils.Green("Successfully applied " + fmt.Sprintf("%d", count) + " migrations"))

	db.Close()
}
