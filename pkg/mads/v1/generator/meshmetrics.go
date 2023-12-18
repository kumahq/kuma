package generator

import (
	"net"
	"strconv"

	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/mads"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
)

func Generate(meshMetricToDataplanes map[*v1alpha1.Conf]*core_mesh.DataplaneResource) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	// this is not a great way of sorting, we could get the policies from the store already sorted
	for meshMetricConf, dataplane := range meshMetricToDataplanes {
		for _, backend := range *meshMetricConf.Backends {
			if backend.Type != v1alpha1.PrometheusBackendType {
				continue
			}

			prometheusEndpoint := backend.Prometheus

			schema := "http"
			if prometheusEndpoint.Tls != nil && prometheusEndpoint.Tls.Mode == v1alpha1.ProvidedTLS {
				schema = "https"
			}

			assignment := &observability_v1.MonitoringAssignment{
				Mesh:    dataplane.Meta.GetMesh(),
				Service: dataplane.Spec.GetIdentifyingService(),
				Targets: []*observability_v1.MonitoringAssignment_Target{{
					Scheme:      schema,
					Name:        dataplane.GetMeta().GetName(),
					Address:     net.JoinHostPort(dataplane.GetIP(), strconv.FormatUint(uint64(prometheusEndpoint.Port), 10)),
					MetricsPath: prometheusEndpoint.Path,
					Labels:      mads.DataplaneLabels(dataplane, []*core_mesh.MeshGatewayResource{}), // what is this mesh gw and why we have labels there?
				}},
			}

			resources = append(resources, &core_xds.Resource{
				Name:     mads.DataplaneAssignmentName(dataplane),
				Resource: assignment,
			})
		}
	}

	return resources, nil
}
