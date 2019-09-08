package config

import (
	kumactl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/cmd"
	"github.com/spf13/cobra"
)

func newConfigControlPlanesCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "control-planes",
		Short: "Manage known Control Planes",
		Long:  `Manage known Control Planes.`,
	}
	// sub-commands
	cmd.AddCommand(newConfigControlPlanesListCmd(pctx))
	cmd.AddCommand(newConfigControlPlanesAddCmd(pctx))
	cmd.AddCommand(newConfigControlPlanesRemoveCmd(pctx))
	cmd.AddCommand(newConfigControlPlanesSwitchCmd(pctx))
	return cmd
}
