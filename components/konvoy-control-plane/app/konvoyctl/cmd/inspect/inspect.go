package inspect

import (
	konvoyctl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output"
	"github.com/spf13/cobra"
)

type inspectContext struct {
	*konvoyctl_cmd.RootContext

	args struct {
		outputFormat string
	}
}

func NewInspectCmd(pctx *konvoyctl_cmd.RootContext) *cobra.Command {
	ctx := &inspectContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect Konvoy resources",
		Long:  `Inspect Konvoy resources.`,
	}
	// flags
	cmd.PersistentFlags().StringVarP(&ctx.args.outputFormat, "output", "o", string(output.TableFormat), konvoyctl_cmd.UsageOptions("Output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	cmd.AddCommand(newInspectDataplanesCmd(ctx))
	return cmd
}
