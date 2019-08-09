package config

import (
	konvoyctl_ctx "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/context"
	"github.com/spf13/cobra"
)

func NewConfigCmd(pctx *konvoyctl_ctx.RootContext) *cobra.Command {
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
