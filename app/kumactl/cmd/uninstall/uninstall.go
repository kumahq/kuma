package uninstall

import (
	"github.com/spf13/cobra"
)

func NewUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall various Kuma components.",
		Long:  `Uninstall various Kuma components.`,
	}

	// sub-commands
	cmd.AddCommand(newUninstallTransparentProxy())

	return cmd
}
