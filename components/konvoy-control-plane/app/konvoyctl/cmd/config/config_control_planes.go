package config

import (
	konvoyctl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/spf13/cobra"
)

func newConfigControlPlanesCmd(pctx *konvoyctl_cmd.RootContext) *cobra.Command {
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
