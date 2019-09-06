package inspect

import (
	kumactl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/output"
	"github.com/spf13/cobra"
)

type inspectContext struct {
	*kumactl_cmd.RootContext

	args struct {
		outputFormat string
	}
}

func NewInspectCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &inspectContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect Kuma resources",
		Long:  `Inspect Kuma resources.`,
	}
	// flags
	cmd.PersistentFlags().StringVarP(&ctx.args.outputFormat, "output", "o", string(output.TableFormat), kumactl_cmd.UsageOptions("Output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	cmd.AddCommand(newInspectDataplanesCmd(ctx))
	return cmd
}
