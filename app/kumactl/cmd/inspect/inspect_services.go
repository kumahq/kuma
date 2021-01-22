package inspect

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

type inspectServicesContext struct {
	*inspectContext

	mesh string
}

func newInspectServicesCmd(pctx *inspectContext) *cobra.Command {
	ctx := inspectServicesContext{
		inspectContext: pctx,
	}
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Inspect Services",
		Long:  `Inspect Services.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := ctx.CurrentServiceOverviewClient()
			if err != nil {
				return err
			}
			insights, err := client.List(context.Background(), ctx.mesh)
			if err != nil {
				return err
			}

			switch format := output.Format(ctx.args.outputFormat); format {
			case output.TableFormat:
				return printServiceInsights(insights, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(insights), cmd.OutOrStdout())
			}
		},
	}
	cmd.PersistentFlags().StringVarP(&ctx.mesh, "mesh", "m", "default", "mesh")

	return cmd
}

func printServiceInsights(overviews *mesh.ServiceOverviewResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{
			"SERVICE",
			"STATUS",
			"DATAPLANES",
		},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(overviews.Items) <= i {
					return nil
				}
				overview := overviews.Items[i]
				return []string{
					overview.Meta.GetName(),                                         // SERVICE
					overview.GetStatus().String(),                                   // STATUS
					fmt.Sprintf("%d/%d", overview.Spec.Online, overview.Spec.Total), // DATAPLANES
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
