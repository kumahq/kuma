package cmd

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output"
	"github.com/spf13/cobra"
)

type getContext struct {
	*rootContext

	args struct {
		outputFormat string
	}
}

func newGetCmd(pctx *rootContext) *cobra.Command {
	ctx := &getContext{rootContext: pctx}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show Konvoy resources",
		Long:  `Show Konvoy resources.`,
	}
	// flags
	cmd.PersistentFlags().StringVarP(&ctx.args.outputFormat, "output", "o", string(output.TableFormat), UsageOptions("Output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	cmd.AddCommand(newGetDataplanesCmd(ctx))
	cmd.AddCommand(newGetMeshesCmd(ctx))
	return cmd
}
