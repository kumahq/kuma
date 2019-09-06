package config

import (
	kumactl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewConfigCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage kumactl config",
		Long:  `Manage kumactl config.`,
	}
	// sub-commands
	cmd.AddCommand(newConfigViewCmd(pctx))
	cmd.AddCommand(newConfigControlPlanesCmd(pctx))
	return cmd
}
