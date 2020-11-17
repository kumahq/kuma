package inspect

import (
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
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
			client, err := ctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			insights := &mesh.ServiceInsightResource{}
			if err := client.Get(context.Background(), insights, store.GetByKey(ctx.mesh, model.NoMesh)); err != nil {
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
				return printer.Print(rest_types.From.Resource(insights), cmd.OutOrStdout())
			}
		},
	}
	cmd.PersistentFlags().StringVarP(&ctx.mesh, "mesh", "m", "default", "mesh")

	return cmd
}

func printServiceInsights(serviceInsight *mesh.ServiceInsightResource, out io.Writer) error {
	data := printers.Table{
		Headers: []string{
			"SERVICE",
			"DATAPLANES",
		},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(serviceInsight.Spec.Services) <= i {
					return nil
				}
				services := []string{}
				for service := range serviceInsight.Spec.Services {
					services = append(services, service)
				}
				sort.Strings(services)

				service := services[i]
				stat := serviceInsight.Spec.Services[service]

				return []string{
					service, // SERVICE
					fmt.Sprintf("%d/%d", stat.Online, stat.Total), // DATAPLANES
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
