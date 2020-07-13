package config

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
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
