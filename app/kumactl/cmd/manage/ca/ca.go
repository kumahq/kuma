package ca

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
)

func NewCaCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ca",
		Short: "Manage certificate authorities",
		Long:  `Manage certificate authorities.`,
	}
	// sub-commands
	cmd.AddCommand(newProvidedCmd(pctx))
	return cmd
}
