package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type HealthCheckConfigurer struct {
	HealthCheck *mesh_core.HealthCheckResource
}

var _ ClusterConfigurer = &HealthCheckConfigurer{}

func (e *HealthCheckConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
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
