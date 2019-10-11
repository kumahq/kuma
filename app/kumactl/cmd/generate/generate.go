package generate

import (
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewGenerateCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate resources",
		Long:  `Generate resources.`,
	}
	// sub-commands
	cmd.AddCommand(NewGenerateDpTokenCmd(pctx))
	return cmd
}
