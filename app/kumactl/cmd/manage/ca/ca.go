package ca

import (
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/spf13/cobra"
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
