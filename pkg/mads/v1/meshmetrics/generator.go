package meshmetrics

import (
	"net"
	"strconv"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/mads"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
)

func Generate(meshMetricToDataplanes map[*v1alpha1.MeshMetricResource][]*core_mesh.DataplaneResource) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource
	assignmentNameToAssignment := map[string]*observability_v1.MonitoringAssignment{}

	for _, targetRefKind := range common_api.OrderInArray {
		for meshMetric, dataplanes := range meshMetricToDataplanes {
			if meshMetric.Spec.TargetRef.Kind != targetRefKind {
				continue
			}
			for _, backend := range *meshMetric.Spec.Default.Backends {
				if backend.Type != v1alpha1.PrometheusBackendType {
					continue
				}

				prometheusEndpoint := backend.Prometheus

				schema := "http"
				if prometheusEndpoint.Tls != nil && prometheusEndpoint.Tls.Mode == v1alpha1.ProvidedTLS {
					schema = "https"
				}

				for _, dataplane := range dataplanes {
					// TODO: could also group by service, and have one assignment per service
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

					assignmentNameToAssignment[mads.DataplaneAssignmentName(dataplane)] = assignment
				}
			}
		}
	}

	for name, assignment := range assignmentNameToAssignment {
		resources = append(resources, &core_xds.Resource{
			Name:     name,
			Resource: assignment,
		})
	}

	return resources, nil
}
