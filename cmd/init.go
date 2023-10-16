package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	_init "github.com/supagrate/cli/internal/init"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize supagrate to work with supabase",
		Long:  `Initialize supagrate to work with supabase`,
		Run: func(cmd *cobra.Command, args []string) {
			_init.Run(afero.NewOsFs())
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
