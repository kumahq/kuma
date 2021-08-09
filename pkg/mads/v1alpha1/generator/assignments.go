package generator

import (
	prom "github.com/prometheus/common/model"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	observability_v1alpha1 "github.com/kumahq/kuma/api/observability/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/mads"
	"github.com/kumahq/kuma/pkg/mads/generator"
)

const (
	// meshLabel is the name of the label that holds the mesh name.
	meshLabel = "mesh"
	// dataplaneLabel is the name of the label that holds the dataplane name.
	dataplaneLabel = "dataplane"
)

var log = core.Log.WithName("mads").WithName("generator")

// MonitoringAssignmentsGenerator knows how to generate MonitoringAssignment
// resources for a given set of Dataplanes.
//
// Beware of the following constraints when it comes to integration with Prometheus:
//
//  1. Prometheus model for all `sd`s except for `file_sd` looks like this:
//
//     // Group is a set of targets with a common label set(production , test, staging etc.).
//     type Group struct {
//         // Targets is a list of targets identified by a label set. Each target is
//         // uniquely identifiable in the group by its address label.
//         Targets []model.LabelSet
//         // Labels is a set of labels that is common across all targets in the group.
//         Labels model.LabelSet
//
//         // Source is an identifier that describes a group of targets.
//         Source string
//     }
//
//     That is why Kuma's MonitoringAssignment was designed to be close to that model.
//
//  2. However, `file_sd` uses different model for reading data from a file:
//
//     struct {
//         Targets []string       `yaml:"targets"`
//         Labels  model.LabelSet `yaml:"labels"`
//     }
//
//     Notice that Targets is just a list of addresses rather than a list of model.LabelSet.
//
//  3. Because of that mismatch, some form of conversion is unavoidable on client side,
//     e.g. inside `kuma-prometheus-sd`
//
//  4. The next component that imposes its constraints is `custom-sd`- adapter
//     (https://github.com/prometheus/prometheus/tree/master/documentation/examples/custom-sd)
//     that is recommended for use by all `file_sd`-based `sd`s.
//
//     This adapter is doing conversion from Prometheus model into `file_sd` model
//     and it expects that `Targets` field has only 1 label - `__address__` -
//     and the rest of the labels must be a part of `Labels` field.
//
//  5. Therefore, we need to convert MonitoringAssignment into a model that `custom-sd` expects.
//     It could happen on server side, it could happen on client side.
//     Given that we're trying to minimize amount of logic on the client side,
//     the choice was made in favour of server side.
//
//  In practice, it means that generated MonitoringAssignment will look the following way:
//
//     name: /meshes/default/dataplanes/backend-01
//     targets:
//     - labels:
//         __address__: 192.168.0.1:8080
//     labels:
//       __scheme__: http
//       __metrics_path__: /metrics
//       job: backend
//       instance: backend-01
//       mesh: default
//       dataplane: backend-01
//       env: prod
//       envs: ,prod,
//       service: backend
//       services: ,backend,
//
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

		prometheusEndpoint, err := dataplane.GetPrometheusEndpoint(mesh)
		if err != nil {
			log.Info("could not get prometheus endpoint from the dataplane", err)
			// does not return error to not break MADS for other dataplanes
			continue
		}

		if prometheusEndpoint == nil {
			// Prometheus metrics are not enabled on that Mesh
			continue
		}

		assignment := &observability_v1alpha1.MonitoringAssignment{
			Name: mads.DataplaneAssignmentName(dataplane),
			Targets: []*observability_v1alpha1.MonitoringAssignment_Target{{
				Labels: g.addressLabel(dataplane, prometheusEndpoint),
			}},
			Labels: g.dataplaneLabels(dataplane, prometheusEndpoint),
		}

		resources = append(resources, &core_xds.Resource{
			Name:     assignment.Name,
			Resource: assignment,
		})
	}

	return resources, nil
}

func (_ MonitoringAssignmentsGenerator) addressLabel(dataplane *core_mesh.DataplaneResource, endpoint *mesh_proto.PrometheusMetricsBackendConfig) map[string]string {
	return map[string]string{
		prom.AddressLabel: mads.Address(dataplane, endpoint),
	}
}

func (g MonitoringAssignmentsGenerator) dataplaneLabels(dataplane *core_mesh.DataplaneResource, endpoint *mesh_proto.PrometheusMetricsBackendConfig) map[string]string {
	labels := mads.DataplaneLabels(dataplane)
	// then, we apply mandatory Prometheus labels on top
	labels[prom.SchemeLabel] = "http"
	labels[prom.MetricsPathLabel] = endpoint.GetPath()
	labels[prom.JobLabel] = dataplane.Spec.GetIdentifyingService()
	labels[prom.InstanceLabel] = dataplane.Meta.GetName()
	labels[meshLabel] = dataplane.Meta.GetMesh()
	labels[dataplaneLabel] = dataplane.Meta.GetName()
	// notice that `service` tag is handled as part of user-defined tags
	return labels
}
