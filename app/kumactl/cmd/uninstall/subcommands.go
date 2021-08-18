// +build !windows

package uninstall

import "github.com/spf13/cobra"

var subcommands = []func() *cobra.Command{
	newUninstallTransparentProxy,
}
