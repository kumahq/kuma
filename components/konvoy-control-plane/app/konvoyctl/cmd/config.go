package cmd

import (
	"github.com/spf13/cobra"
)

func newConfigCmd(pctx *rootContext) *cobra.Command {
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
