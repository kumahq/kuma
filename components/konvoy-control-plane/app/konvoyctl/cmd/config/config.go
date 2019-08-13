package config

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewConfigCmd(pctx *cmd.RootContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: "Manage konvoyctl config",
		Long:  `Manage konvoyctl config.`,
	}
	// sub-commands
	command.AddCommand(newConfigViewCmd(pctx))
	command.AddCommand(newConfigControlPlanesCmd(pctx))
	return command
}
