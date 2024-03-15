package generator

import (
	"net"
	"strconv"

	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/mads"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

const DefaultKumaClientId = "_kuma-default-client"

func Generate(meshMetricToDataplane map[*v1alpha1.Conf]*core_mesh.DataplaneResource, clientId string) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	for meshMetricConf, dataplane := range meshMetricToDataplane {
		for _, backend := range *meshMetricConf.Backends {
			if backend.Type != v1alpha1.PrometheusBackendType {
				continue
			}
			prometheusEndpoint := backend.Prometheus

			if pointer.DerefOr(prometheusEndpoint.ClientId, DefaultKumaClientId) != clientId {
				continue
			}

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
