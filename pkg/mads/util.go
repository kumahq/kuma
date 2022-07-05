package mads

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	prom_util "github.com/prometheus/prometheus/util/strutil"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

func IndexMeshes(meshes []*core_mesh.MeshResource) map[string]*core_mesh.MeshResource {
	index := make(map[string]*core_mesh.MeshResource)
	for _, mesh := range meshes {
		index[mesh.Meta.GetName()] = mesh
	}
	return index
}

func Address(dataplane *core_mesh.DataplaneResource, endpoint *mesh_proto.PrometheusMetricsBackendConfig) string {
	// TODO: handle a case where Dataplane's IP is unknown
	// For now, we export such a Dataplane with an empty IP address, so that the error state will be at least visible on the Prometheus side
	return net.JoinHostPort(dataplane.GetIP(), strconv.FormatUint(uint64(endpoint.GetPort()), 10))
}

func MultiValue(values []string) string {
	// Although looks weird, it's actually a recommended way to represent multi-values in Prometheus.
	// It's meant to simplify greatly user-defined queries, e.g. just `,value,` instead of a full-featured regex.
	return "," + strings.Join(values, ",") + ","
}

func DataplaneLabels(dataplane *core_mesh.DataplaneResource, meshGateways []*core_mesh.MeshGatewayResource) map[string]string {
	labels := map[string]string{}
	// first, we copy user-defined tags
	tags := dataplane.Spec.TagSet()
	for _, key := range tags.Keys() {
		values := tags.Values(key)
		value := ""
		if len(values) > 0 {
			value = values[0]
		}
		// while in general case a tag might have multiple values, we want to optimize for a single-value scenario
		labels[prom_util.SanitizeLabelName(key)] = value
		// additionally, we also support a multi-value scenario by automatically pluralizing label name,
		// e.g. `service => services`, `version => versions`, etc.
		// if it happens that a user defined both `service` and `services` tags,
		// user-defined `services` tag will override auto-generated one (since keys are iterated in a sorted order)
		plural := fmt.Sprintf("%ss", key)
		labels[prom_util.SanitizeLabelName(plural)] = MultiValue(values)
	}
	// then, we turn name extensions into labels
	for key, value := range dataplane.GetMeta().GetNameExtensions() {
		labels[prom_util.SanitizeLabelName(key)] = value
	}

	if dataplane.Spec.IsBuiltinGateway() {
		gateway := topology.SelectGateway(meshGateways, dataplane.Spec.Matches)
		name := ""
		if gateway != nil {
			name = gateway.GetMeta().GetName()
		}
		labels["kuma_io_mesh_gateway"] = name
	}

	return labels
}

func DataplaneAssignmentName(dataplane *core_mesh.DataplaneResource) string {
	// unique name, e.g. REST API uri
	return fmt.Sprintf("/meshes/%s/dataplanes/%s", dataplane.Meta.GetMesh(), dataplane.Meta.GetName())
}
