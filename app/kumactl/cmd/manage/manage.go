package manage

import (
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/cmd/manage/ca"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
)

func NewManageCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manage",
		Short: "Manage certificate authorities, etc",
		Long:  `Manage certificate authorities, etc.`,
	}
	// sub-commands
	cmd.AddCommand(ca.NewCaCmd(pctx))
	return cmd
}
