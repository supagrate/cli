package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/supagrate/cli/internal/seed/apply"
	"github.com/supagrate/cli/internal/seed/new"
	"github.com/supagrate/cli/internal/utils"
)

// seedCmd represents the seed command
var (
	seedCmd = &cobra.Command{
		Use:   "seed",
		Short: "Manage database seeds",
		Long:  `Manage database seeds`,
	}

	seedApplyCmd = &cobra.Command{
		Use:     "apply [flags] <seed name>",
		Example: `supagrate seed apply insert_countries`,
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Short:   "Apply a seed to your database",
		Long:    `Apply a seed to your database`,
		Run: func(cmd *cobra.Command, args []string) {
			apply.Run(cmd, args[0], afero.NewOsFs())
		},
	}

	seedNewCmd = &cobra.Command{
		Use:     "new [flags] <seed name>",
		Example: `supagrate seed new insert_countries`,
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Short:   "Generate a new database seed",
		Long:    `Generate a new database seed`,
		Run: func(cmd *cobra.Command, args []string) {
			new.Run(args[0], afero.NewOsFs())
		},
	}
)

func init() {
	utils.UseDBFlags(seedApplyCmd)

	seedCmd.AddCommand(seedApplyCmd)
	seedCmd.AddCommand(seedNewCmd)

	rootCmd.AddCommand(seedCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// seedCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// seedCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
