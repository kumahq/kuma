package generator

import (
	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
	"github.com/kumahq/kuma/pkg/core"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/mads"
	"github.com/kumahq/kuma/pkg/mads/generator"
)

var log = core.Log.WithName("mads").WithName("v1").WithName("generator")

// MonitoringAssignmentsGenerator knows how to generate MonitoringAssignment
// resources for a given set of Dataplanes.
type MonitoringAssignmentsGenerator struct {
}

// Generate implements mads.ResourceGenerator
func (g MonitoringAssignmentsGenerator) Generate(args generator.Args) ([]*core_xds.Resource, error) {
	meshIndex := mads.IndexMeshes(args.Meshes)

	resources := make([]*core_xds.Resource, 0, len(args.Dataplanes))
	for _, dataplane := range args.Dataplanes {
		mesh, exist := meshIndex[dataplane.Meta.GetMesh()]
		if !exist {
			// might be the case when the entire mesh is in the process of being deleted
			continue
		}

		prometheusEndpoint, err := dataplane.GetPrometheusConfig(mesh)
		if err != nil {
			log.Info("could not get prometheus endpoint from the dataplane", err)
			// does not return error to not break MADS for other dataplanes
			continue
		}

		if prometheusEndpoint == nil {
			// Prometheus metrics are not enabled on that Mesh
			continue
		}

		// TODO: could also group by service, and have one assignment per service
		assignment := &observability_v1.MonitoringAssignment{
			Mesh:    dataplane.Meta.GetMesh(),
			Service: dataplane.Spec.GetIdentifyingService(),
			Targets: []*observability_v1.MonitoringAssignment_Target{{
				Scheme:      "http",
				Name:        dataplane.GetMeta().GetName(),
				Address:     mads.Address(dataplane, prometheusEndpoint),
				MetricsPath: prometheusEndpoint.GetPath(),
				Labels:      mads.DataplaneLabels(dataplane, args.MeshGateways),
			}},
		}

		resources = append(resources, &core_xds.Resource{
			Name:     mads.DataplaneAssignmentName(dataplane),
			Resource: assignment,
		})
	}

	return resources, nil
}
