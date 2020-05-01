package generator

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"

	prom "github.com/prometheus/common/model"
	prom_util "github.com/prometheus/prometheus/util/strutil"
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

func (g MonitoringAssignmentsGenerator) Generate(args Args) ([]*core_xds.Resource, error) {
	meshIndex := g.indexMeshes(args.Meshes)

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

		assignment := &observability_proto.MonitoringAssignment{
			Name: g.assignmentName(dataplane),
			Targets: []*observability_proto.MonitoringAssignment_Target{{
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

func (_ MonitoringAssignmentsGenerator) indexMeshes(meshes []*mesh_core.MeshResource) map[string]*mesh_core.MeshResource {
	index := make(map[string]*mesh_core.MeshResource)
	for _, mesh := range meshes {
		index[mesh.Meta.GetName()] = mesh
	}
	return index
}

func (_ MonitoringAssignmentsGenerator) assignmentName(dataplane *mesh_core.DataplaneResource) string {
	// unique name, e.g. REST API uri
	return fmt.Sprintf("/meshes/%s/dataplanes/%s", dataplane.Meta.GetMesh(), dataplane.Meta.GetName())
}

func (_ MonitoringAssignmentsGenerator) addressLabel(dataplane *mesh_core.DataplaneResource, endpoint *mesh_proto.PrometheusMetricsBackendConfig) map[string]string {
	// TODO(yskopets): handle a case where Dataplane's IP is unknown
	// For now, we export such a Dataplane with an empty IP address, so that the error state will be at least visible on the Prometheus side
	return map[string]string{
		prom.AddressLabel: net.JoinHostPort(dataplane.GetIP(), strconv.FormatUint(uint64(endpoint.GetPort()), 10)),
	}
}

func (g MonitoringAssignmentsGenerator) dataplaneLabels(dataplane *mesh_core.DataplaneResource, endpoint *mesh_proto.PrometheusMetricsBackendConfig) map[string]string {
	labels := map[string]string{}
	// first, we copy user-defined tags
	tags := dataplane.Spec.Tags()
	for _, key := range tags.Keys() {
		values := tags.Values(key)
		value := ""
		if len(values) > 0 {
			value = values[0]
		}
		// while in general case a tag might have multiple values, we want to omptimize for a single-value scenario
		labels[prom_util.SanitizeLabelName(key)] = value
		// additionally, we also support a multi-value scenario by automatically pluralizing label name,
		// e.g. `service => services`, `version => versions`, etc.
		// if it happens that a user defined both `service` and `services` tags,
		// user-defined `services` tag will override auto-generated one (since keys are iterated in a sorted order)
		plural := fmt.Sprintf("%ss", key)
		labels[prom_util.SanitizeLabelName(plural)] = g.multiValue(values)
	}
	// then, we turn name extensions into labels
	for key, value := range dataplane.GetMeta().GetNameExtensions() {
		labels[prom_util.SanitizeLabelName(key)] = value
	}
	// then, we apply mandatory labels on top
	labels[prom.SchemeLabel] = "http"
	labels[prom.MetricsPathLabel] = endpoint.GetPath()
	labels[prom.JobLabel] = dataplane.Spec.GetIdentifyingService()
	labels[prom.InstanceLabel] = dataplane.Meta.GetName()
	labels[meshLabel] = dataplane.Meta.GetMesh()
	labels[dataplaneLabel] = dataplane.Meta.GetName()
	// notice that `service` tag is handled as part of user-defined tags
	return labels
}

func (_ MonitoringAssignmentsGenerator) multiValue(values []string) string {
	// Although looks weird, it's actually a recommended way to represent multi-values in Prometheus.
	// It's meant to simplify greatly user-defined queries, e.g. just `,value,` instead of a full-featured regex.
	return "," + strings.Join(values, ",") + ","
}
