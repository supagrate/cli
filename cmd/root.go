package cmd

import (
	"io"
	"os"

	cc "github.com/ivanpirog/coloredcobra"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/supagrate/cli/internal/utils"
)

// rootCmd represents the base command when called without any subcommands
var (
	verbosity string
	rootCmd   = &cobra.Command{
		Use:     "supagrate",
		Short:   "Supagrate CLI " + utils.Version,
		Version: utils.Version,
		Long:    `Supagrate is a CLI library meant to improve the database migration experience with Supabase.`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cc.Init(&cc.Config{
		RootCmd:  rootCmd,
		Headings: cc.HiWhite + cc.Bold,
		Commands: cc.HiYellow,
		Example:  cc.Italic,
		ExecName: cc.HiGreen + cc.Bold,
	})
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here is where we define the PreRun func, using the verbose flag value
	// We use the standard output for logs.
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := setupLogs(os.Stdout, verbosity); err != nil {
			return err
		}
		return nil
	}

	// Here is where we bind the verbose flag
	// Default value is the warn level
	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", logrus.WarnLevel.String(), "Record level (debug, info, warn, error, fatal, panic")

	rootCmd.SetVersionTemplate("{{.Version}}\n")
}

func setupLogs(out io.Writer, level string) error {
	logrus.SetOutput(out)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	return nil
}
