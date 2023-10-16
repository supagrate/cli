package init

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/supagrate/cli/internal/utils"
)

func Run(fs afero.Fs) {
	//alreadyInitialized()

	//accessToken, err := utils.GetSupabaseToken(fs)

	//if err != nil {
	//	logrus.Fatal(err)
	//	os.Exit(1)
	//}

	//items, err := projects.ReadProjects(accessToken)

	createDirectory(fs)
}

func createDirectory(fs afero.Fs) {
	err := utils.MkdirIfNotExist(fs, utils.SupagrateDirectory)

	if err != nil {
		logrus.Panic(err)
		os.Exit(1)
	}
}

func alreadyInitialized() {
	if _, err := os.Stat(utils.SupagrateDirectory); !os.IsNotExist(err) {
		fmt.Println(utils.Green("Supagrate is already initialized."))
		os.Exit(1)
	}
}

func EnsureInitialization() {
	if _, err := os.Stat(utils.SupagrateDirectory); os.IsNotExist(err) {
		fmt.Println(utils.Red("Supagrate is not initialized. Please run `supagrate init` first."))
		os.Exit(1)
	}
}
