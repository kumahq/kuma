package generate

import (
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewGenerateCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate resources, tokens etc",
		Long:  `Generate resources, tokens etc.`,
	}
	// sub-commands
	cmd.AddCommand(NewGenerateDataplaneTokenCmd(pctx))
	return cmd
}
