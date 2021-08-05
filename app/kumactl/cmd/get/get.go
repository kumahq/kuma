package get

import (
	"github.com/spf13/cobra"

	get_context "github.com/kumahq/kuma/app/kumactl/cmd/get/context"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/entities"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
)

func NewGetCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show Kuma resources",
		Long:  `Show Kuma resources.`,
	}
	// flags
	cmd.PersistentFlags().StringVarP(&pctx.GetContext.Args.OutputFormat, "output", "o", string(output.TableFormat), kuma_cmd.UsageOptions("output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	for _, cmdInst := range entities.All {
		cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, cmdInst.Plural, cmdInst.ResourceType), &pctx.ListContext))
		cmd.AddCommand(NewGetResourceCmd(pctx, cmdInst.Singular, cmdInst.ResourceType))
	}
	return cmd
}

func WithPaginationArgs(cmd *cobra.Command, ctx *get_context.ListContext) *cobra.Command {
	cmd.PersistentFlags().IntVarP(&ctx.Args.Size, "size", "", 0, "maximum number of elements to return")
	cmd.PersistentFlags().StringVarP(&ctx.Args.Offset, "offset", "", "", "the offset that indicates starting element of the resources list to retrieve")
	return cmd
}
