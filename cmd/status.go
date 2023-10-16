package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/supagrate/cli/internal/status"
	"github.com/supagrate/cli/internal/utils"
)

// statusCmd represents the status command
var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "See your database migration status",
		Long:  `See your database migration status`,
		Run: func(cmd *cobra.Command, args []string) {
			status.Run(cmd, afero.NewOsFs())
		},
	}
)

func init() {
	utils.UseDBFlags(statusCmd)

	rootCmd.AddCommand(statusCmd)
}
