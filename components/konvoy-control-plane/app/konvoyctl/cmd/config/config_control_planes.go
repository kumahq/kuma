package config

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/spf13/cobra"
)

func newConfigControlPlanesCmd(pctx *cmd.RootContext) *cobra.Command {
	commandd := &cobra.Command{
		Use:   "control-planes",
		Short: "Manage known Control Planes",
		Long:  `Manage known Control Planes.`,
	}
	// sub-commands
	commandd.AddCommand(newConfigControlPlanesListCmd(pctx))
	commandd.AddCommand(newConfigControlPlanesAddCmd(pctx))
	return commandd
}
