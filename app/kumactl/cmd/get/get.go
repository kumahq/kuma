package get

import (
	"github.com/spf13/cobra"

	get_context "github.com/kumahq/kuma/app/kumactl/cmd/get/context"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/register"
)

func NewGetCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	register.RegisterGatewayTypes() // allow applying experimental Gateway types

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Show Kuma resources",
		Long:  `Show Kuma resources.`,
	}
	getCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := kumactl_cmd.RunParentPreRunE(getCmd, args); err != nil {
			return err
		}
		if err := pctx.CheckServerVersionCompatibility(); err != nil {
			cmd.PrintErrln(err)
		}
		return nil
	}
	// flags
	getCmd.PersistentFlags().StringVarP(&pctx.GetContext.Args.OutputFormat, "output", "o", string(output.TableFormat), kuma_cmd.UsageOptions("output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	for _, cmdInst := range pctx.Runtime.Registry.ObjectDescriptors(model.HasKumactlEnabled()) {
		getCmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, cmdInst), &pctx.ListContext))
		getCmd.AddCommand(NewGetResourceCmd(pctx, cmdInst))
	}
	return getCmd
}

func WithPaginationArgs(cmd *cobra.Command, ctx *get_context.ListContext) *cobra.Command {
	cmd.PersistentFlags().IntVarP(&ctx.Args.Size, "size", "", 0, "maximum number of elements to return")
	cmd.PersistentFlags().StringVarP(&ctx.Args.Offset, "offset", "", "", "the offset that indicates starting element of the resources list to retrieve")
	return cmd
}
