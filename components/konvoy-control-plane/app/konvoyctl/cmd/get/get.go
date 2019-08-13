package get

import (
	konvoyctl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output"
	"github.com/spf13/cobra"
)

type getContext struct {
	*konvoyctl_cmd.RootContext

	args struct {
		outputFormat string
	}
}

func NewGetCmd(pctx *konvoyctl_cmd.RootContext) *cobra.Command {
	ctx := &getContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show Konvoy resources",
		Long:  `Show Konvoy resources.`,
	}
	// flags
	cmd.PersistentFlags().StringVarP(&ctx.args.outputFormat, "output", "o", string(output.TableFormat), konvoyctl_cmd.UsageOptions("Output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	cmd.AddCommand(newGetDataplanesCmd(ctx))
	cmd.AddCommand(newGetMeshesCmd(ctx))
	return cmd
}
