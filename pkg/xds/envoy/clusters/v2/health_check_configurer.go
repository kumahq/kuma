package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type HealthCheckConfigurer struct {
	HealthCheck *mesh_core.HealthCheckResource
}

var _ ClusterConfigurer = &HealthCheckConfigurer{}

func (e *HealthCheckConfigurer) Configure(cluster *envoy_api.Cluster) error {
	if e.HealthCheck == nil || e.HealthCheck.Spec.Conf == nil {
		return nil
	}
	activeChecks := e.HealthCheck.Spec.Conf
	cluster.HealthChecks = append(cluster.HealthChecks, &envoy_core.HealthCheck{
		HealthChecker: &envoy_core.HealthCheck_TcpHealthCheck_{
			TcpHealthCheck: &envoy_core.HealthCheck_TcpHealthCheck{},
		},
		Interval:           activeChecks.Interval,
		Timeout:            activeChecks.Timeout,
		UnhealthyThreshold: &wrappers.UInt32Value{Value: activeChecks.UnhealthyThreshold},
		HealthyThreshold:   &wrappers.UInt32Value{Value: activeChecks.HealthyThreshold},
	})
	return nil
}
