package config

import (
	konvoyctl_ctx "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/context"
	"github.com/spf13/cobra"
)

func newConfigControlPlanesCmd(pctx *konvoyctl_ctx.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "control-planes",
		Short: "Manage known Control Planes",
		Long:  `Manage known Control Planes.`,
	}
	// sub-commands
	cmd.AddCommand(newConfigControlPlanesListCmd(pctx))
	cmd.AddCommand(newConfigControlPlanesAddCmd(pctx))
	return cmd
}
