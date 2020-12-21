package uninstall

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

func NewUninstallCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall various Kuma components.",
		Long:  `Uninstall various Kuma components.`,
	}
	// sub-commands
	cmd.AddCommand(newUninstallTransparentProxy())
	return cmd
}
