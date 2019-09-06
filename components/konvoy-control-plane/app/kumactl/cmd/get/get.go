package get

import (
	kumactl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/output"
	"github.com/spf13/cobra"
)

type getContext struct {
	*kumactl_cmd.RootContext

	args struct {
		outputFormat string
	}
}

func NewGetCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &getContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show Kuma resources",
		Long:  `Show Kuma resources.`,
	}
	// flags
	cmd.PersistentFlags().StringVarP(&ctx.args.outputFormat, "output", "o", string(output.TableFormat), kumactl_cmd.UsageOptions("Output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	cmd.AddCommand(newGetMeshesCmd(ctx))
	cmd.AddCommand(newGetProxyTemplatesCmd(ctx))
	cmd.AddCommand(newGetTrafficPermissionsCmd(ctx))
	return cmd
}
