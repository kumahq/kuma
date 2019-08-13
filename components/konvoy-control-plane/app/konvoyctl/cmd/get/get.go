package get

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output"
	"github.com/spf13/cobra"
)

type getContext struct {
	*cmd.RootContext

	args struct {
		outputFormat string
	}
}

func NewGetCmd(pctx *cmd.RootContext) *cobra.Command {
	ctx := &getContext{RootContext: pctx}
	command := &cobra.Command{
		Use:   "get",
		Short: "Show Konvoy resources",
		Long:  `Show Konvoy resources.`,
	}
	// flags
	command.PersistentFlags().StringVarP(&ctx.args.outputFormat, "output", "o", string(output.TableFormat), cmd.UsageOptions("Output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	command.AddCommand(newGetDataplanesCmd(ctx))
	command.AddCommand(newGetMeshesCmd(ctx))
	return command
}
