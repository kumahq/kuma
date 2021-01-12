package inspect

import (
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
			insights := &mesh.ServiceInsightResourceList{}
			if err := client.List(context.Background(), insights, store.ListByMesh(ctx.mesh)); err != nil {
				return err
			}

			serviceInsight := &mesh.ServiceInsightResource{
				Spec: &mesh_proto.ServiceInsight{
					Services: map[string]*mesh_proto.ServiceInsight_DataplaneStat{},
				},
			}
			if len(insights.Items) != 0 {
				serviceInsight = insights.Items[0]
			}

			switch format := output.Format(ctx.args.outputFormat); format {
			case output.TableFormat:
				return printServiceInsights(serviceInsight, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(serviceInsight), cmd.OutOrStdout())
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
