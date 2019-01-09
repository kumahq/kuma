package generator

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
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

// MonitoringAssignmentsGenerator knows how to generate MonitoringAssignment
// resources for a given set of Dataplanes.
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

		prometheusEndpoint := dataplane.GetPrometheusEndpoint(mesh)
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

func (_ MonitoringAssignmentsGenerator) addressLabel(dataplane *mesh_core.DataplaneResource, endpoint *mesh_proto.Metrics_Prometheus) map[string]string {
	// TODO(yskopets): handle a case where Dataplane's IP is unknown
	// For now, we export such a Dataplane with an empty IP address, so that the error state will be at least visible on the Prometheus side
	return map[string]string{
		prom.AddressLabel: net.JoinHostPort(dataplane.GetIP(), strconv.FormatUint(uint64(endpoint.GetPort()), 10)),
	}
}

func (g MonitoringAssignmentsGenerator) dataplaneLabels(dataplane *mesh_core.DataplaneResource, endpoint *mesh_proto.Metrics_Prometheus) map[string]string {
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
