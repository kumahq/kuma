package cmd

import (
	"github.com/spf13/cobra"
)

func newConfigControlPlanesCmd(pctx *rootContext) *cobra.Command {
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
