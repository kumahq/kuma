package inspect

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/v3/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
)

func newInspectMeshesCmd(pctx *cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meshes",
		Short: "Inspect Meshes",
		Long:  `Inspect Meshes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			insights := &mesh.MeshInsightResourceList{}
			if err := client.List(cmd.Context(), insights); err != nil {
				return err
			}

			format := output.Format(pctx.InspectContext.Args.OutputFormat)
			return printers.GenericPrint(format, insights, meshInsightTable, cmd.OutOrStdout())
		},
	}
	return cmd
}

var meshInsightTable = printers.Table{
	Headers: []string{
		"MESH",
		"DATAPLANES",
		"EXTERNAL SERVICES",
	},
	RowForItem: func(i int, container any) ([]string, error) {
		meshInsights := container.(*mesh.MeshInsightResourceList)
		if len(meshInsights.Items) <= i {
			return nil, nil
		}
		meta := meshInsights.Items[i].Meta
		meshInsight := meshInsights.Items[i].Spec

		var es uint32
		if stat, ok := meshInsight.Policies[string(mesh.ExternalServiceType)]; ok {
			es = stat.Total
		}

		return []string{
			meta.GetName(), // MESH
			fmt.Sprintf("%d/%d", meshInsight.Dataplanes.Online, meshInsight.Dataplanes.Total), // DATAPLANES
			table.Number(es), // EXTERNAL SERVICES
		}, nil
	},
}
