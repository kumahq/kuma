package inspect

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
		"TRAFFIC PERMISSIONS",
		"TRAFFIC ROUTES",
		"CIRCUIT BREAKERS",
		"HEALTH CHECKS",
		"FAULT INJECTIONS",
		"EXTERNAL SERVICES",
		"TRAFFIC TRACES",
		"TRAFFIC LOGS",
		"PROXY TEMPLATES",
		"RATE LIMITS",
	},
	RowForItem: func(i int, container interface{}) ([]string, error) {
		meshInsights := container.(*mesh.MeshInsightResourceList)
		if len(meshInsights.Items) <= i {
			return nil, nil
		}
		meta := meshInsights.Items[i].Meta
		meshInsight := meshInsights.Items[i].Spec

		var tp uint32
		if stat, ok := meshInsight.Policies[string(mesh.TrafficPermissionType)]; ok {
			tp = stat.Total
		}

		var tr uint32
		if stat, ok := meshInsight.Policies[string(mesh.TrafficRouteType)]; ok {
			tr = stat.Total
		}

		var cb uint32
		if stat, ok := meshInsight.Policies[string(mesh.CircuitBreakerType)]; ok {
			cb = stat.Total
		}

		var hc uint32
		if stat, ok := meshInsight.Policies[string(mesh.HealthCheckType)]; ok {
			hc = stat.Total
		}

		var fi uint32
		if stat, ok := meshInsight.Policies[string(mesh.FaultInjectionType)]; ok {
			fi = stat.Total
		}

		var es uint32
		if stat, ok := meshInsight.Policies[string(mesh.ExternalServiceType)]; ok {
			es = stat.Total
		}

		var tt uint32
		if stat, ok := meshInsight.Policies[string(mesh.TrafficTraceType)]; ok {
			tt = stat.Total
		}

		var tl uint32
		if stat, ok := meshInsight.Policies[string(mesh.TrafficLogType)]; ok {
			tl = stat.Total
		}

		var pt uint32
		if stat, ok := meshInsight.Policies[string(mesh.ProxyTemplateType)]; ok {
			pt = stat.Total
		}

		var rl uint32
		if stat, ok := meshInsight.Policies[string(mesh.RateLimitType)]; ok {
			rl = stat.Total
		}

		return []string{
			meta.GetName(), // MESH
			fmt.Sprintf("%d/%d", meshInsight.Dataplanes.Online, meshInsight.Dataplanes.Total), // DATAPLANES
			table.Number(tp), // TRAFFIC PERMISSIONS
			table.Number(tr), // TRAFFIC ROUTES
			table.Number(cb), // CIRCUIT BREAKERS
			table.Number(hc), // HEALTH CHECKS
			table.Number(fi), // FAULT INJECTIONS
			table.Number(es), // EXTERNAL SERVICES
			table.Number(tt), // TRAFFIC TRACES
			table.Number(tl), // TRAFFIC LOGS
			table.Number(pt), // PROXY TEMPLATES
			table.Number(rl), // RATE LIMITS
		}, nil
	},
}
