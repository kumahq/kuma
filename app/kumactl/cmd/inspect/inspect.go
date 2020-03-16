package inspect

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/output"
	kuma_cmd "github.com/Kong/kuma/pkg/cmd"
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
	cmd.PersistentFlags().StringVarP(&ctx.args.outputFormat, "output", "o", string(output.TableFormat), kuma_cmd.UsageOptions("output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	cmd.AddCommand(newInspectDataplanesCmd(ctx))
	return cmd
}
