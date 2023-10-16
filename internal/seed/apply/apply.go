package apply

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	_init "github.com/supagrate/cli/internal/init"
	"github.com/supagrate/cli/internal/utils"
	"github.com/supagrate/cli/seeds"
)

func Run(cmd *cobra.Command, seedName string, fs afero.Fs) {
	_init.EnsureInitialization()
	utils.UseDBEnvironmentVariables(cmd)

	connection := utils.Connection{
		Host:     cmd.Flag("db-host").Value.String(),
		Port:     cmd.Flag("db-port").Value.String(),
		User:     cmd.Flag("db-user").Value.String(),
		Password: cmd.Flag("db-password").Value.String(),
		Name:     cmd.Flag("db-name").Value.String(),
	}

	db := utils.ConnectDatabase(connection)

	// Find the seed
	seed := seeds.FindSeed(fs, seedName)

	// Apply it
	seed.Apply(db)

	fmt.Println(utils.Green("Seed applied: ") + seed.FileName)

	db.Close()
}
