package new

import (
	"fmt"

	"github.com/spf13/afero"

	_init "github.com/supagrate/cli/internal/init"
	"github.com/supagrate/cli/internal/utils"
	"github.com/supagrate/cli/seeds"
)

func Run(seedName string, fs afero.Fs) {
	_init.EnsureInitialization()

	seeds.EnsureSeedsDirectoryExists(fs)

	path := seeds.GenerateSeedFile(fs, seedName)

	fmt.Println(utils.Green("New seed created at ") + path)
}
