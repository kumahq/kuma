package delete

import (
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/spf13/cobra"
)

type deleteContext struct {
	*kumactl_cmd.RootContext
}

func NewDeleteCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &deleteContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Kuma resources",
		Long:  `Delete Kuma resources.`,
	}

	// sub-commands
	cmd.AddCommand(newDeleteMeshCmd(ctx))
	return cmd
}
