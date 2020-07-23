package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

func HealthCheck(healthCheck *mesh_core.HealthCheckResource) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&healthCheckConfigurer{
			healthCheck: healthCheck,
		})
	})
}

type healthCheckConfigurer struct {
	healthCheck *mesh_core.HealthCheckResource
}

func (e *healthCheckConfigurer) Configure(cluster *envoy_api.Cluster) error {
	if e.healthCheck == nil || e.healthCheck.Spec.Conf == nil {
		return nil
	}
	activeChecks := e.healthCheck.Spec.Conf
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
