package uninstall

import (
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewUninstallCmd(root *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall various Kuma components.",
		Long:  `Uninstall various Kuma components.`,
	}

	// sub-commands
	cmd.AddCommand(newUninstallTransparentProxy())
	cmd.AddCommand(newUninstallEbpf(root))

	return cmd
}
