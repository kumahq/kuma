package config

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

func NewConfigCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Modify kumactl config files",
		Long: `Modify kumactl config files.
This set of commands enable you to manage connection to different control-planes.
It manipulates the configuration file stored at the value of '--config-file' ('~/.kumactl/config' by default).

These commands do not interact with a running control-plane except to verify the connection when adding a control-plane

For example to add a control-plane to the configuration file:
	kumactl config control-planes add --name=example --address=https://kuma-control-plane:5681
`,
	}
	// sub-commands
	cmd.AddCommand(newConfigViewCmd(pctx))
	cmd.AddCommand(newConfigControlPlanesCmd(pctx))
	return cmd
}
