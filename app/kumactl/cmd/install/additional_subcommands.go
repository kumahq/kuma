// +build !windows

package install

import (
	"github.com/spf13/cobra"
)

var additionalSubcommands = []func() *cobra.Command{
	newInstallTransparentProxy,
}
