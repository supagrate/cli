package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/supagrate/cli/internal/migrate/down"
	"github.com/supagrate/cli/internal/migrate/new"
	"github.com/supagrate/cli/internal/migrate/reset"
	"github.com/supagrate/cli/internal/migrate/up"
	"github.com/supagrate/cli/internal/utils"
)

// migrateCmd represents the migrate command
var (
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Manage database migrations",
		Long:  `Manage database migrations`,
	}

	migrateDownCmd = &cobra.Command{
		Use:   "down [flags]",
		Short: "Rollback migration from your database",
		Long:  `Rollback migration from your database`,
		Run: func(cmd *cobra.Command, args []string) {
			down.Run(cmd, afero.NewOsFs())
		},
	}

	migrateNewCmd = &cobra.Command{
		Use:     "new [flags] <migration name>",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Short:   "Generate a new migration",
		Long:    `Generate a new migration`,
		Example: `supagrate migrate new create_users_table`,
		Run: func(cmd *cobra.Command, args []string) {
			new.Run(args[0], afero.NewOsFs())
		},
	}

	migrateResetCmd = &cobra.Command{
		Use:   "reset [flags]",
		Short: "Reset your database and apply migrations again",
		Long:  `Reset your database and apply migrations again`,
		Run: func(cmd *cobra.Command, args []string) {
			reset.Run(cmd, afero.NewOsFs())
		},
	}

	migrateUpCmd = &cobra.Command{
		Use:   "up [flags]",
		Short: "Apply migrations to your database",
		Long:  `Apply migrations to your database`,
		Run: func(cmd *cobra.Command, args []string) {
			up.Run(cmd, afero.NewOsFs())
		},
	}
)

func init() {
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateNewCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateUpCmd)

	utils.UseDBFlags(migrateCmd)

	migrateUpCmd.Flags().StringP("count", "c", "all", "Number of migrations to apply")
	migrateDownCmd.Flags().StringP("count", "c", "all", "Number of migrations to rollback")

	rootCmd.AddCommand(migrateCmd)
}
