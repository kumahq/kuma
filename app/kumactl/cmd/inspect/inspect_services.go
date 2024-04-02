package inspect

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type inspectServicesContext struct {
	mesh string
}

func newInspectServicesCmd(pctx *cmd.RootContext) *cobra.Command {
	ctx := inspectServicesContext{}
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Inspect Services",
		Long:  `Inspect Services.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentServiceOverviewClient()
			if err != nil {
				return err
			}
			insights, err := client.List(context.Background(), ctx.mesh)
			if err != nil {
				return err
			}

			format := output.Format(pctx.InspectContext.Args.OutputFormat)
			return printers.GenericPrint(format, insights, inspectServiceTable, cmd.OutOrStdout())
		},
	}
	cmd.PersistentFlags().StringVarP(&ctx.mesh, "mesh", "m", "default", "mesh")

	return cmd
}

var inspectServiceTable = printers.Table{
	Headers: []string{
		"SERVICE",
		"STATUS",
		"DATAPLANES",
	},
	RowForItem: func(i int, container interface{}) ([]string, error) {
		overviews := container.(*mesh.ServiceOverviewResourceList)
		if len(overviews.Items) <= i {
			return nil, nil
		}
		overview := overviews.Items[i]
		return []string{
			overview.Meta.GetName(),    // SERVICE
			overview.Status().String(), // STATUS
			fmt.Sprintf("%d/%d", overview.Spec.Dataplanes.Online, overview.Spec.Dataplanes.Total), // DATAPLANES
		}, nil
	},
}
