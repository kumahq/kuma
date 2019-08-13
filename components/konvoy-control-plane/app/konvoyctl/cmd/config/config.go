package config

import (
	konvoyctl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewConfigCmd(pctx *konvoyctl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage konvoyctl config",
		Long:  `Manage konvoyctl config.`,
	}
	// sub-commands
	cmd.AddCommand(newConfigViewCmd(pctx))
	cmd.AddCommand(newConfigControlPlanesCmd(pctx))
	return cmd
}
