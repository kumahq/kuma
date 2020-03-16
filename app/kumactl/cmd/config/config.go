package config

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
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
