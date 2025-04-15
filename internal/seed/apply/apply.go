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

	connectionString := cmd.Flag("connection").Value.String()
	var conn *utils.Connection
	if connectionString != "" {
		conn = &utils.Connection{Connection: utils.ParseConnectionString(connectionString)}
	} else {
		conn = &utils.Connection{Connection: nil}
	}
	db := utils.ConnectDatabase(*conn)

	// Find the seed
	seed := seeds.FindSeed(fs, seedName)

	// Apply it
	seed.Apply(db)

	fmt.Println(utils.Green("Seed applied: ") + seed.FileName)

	db.Close()
}
