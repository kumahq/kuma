package clusters

import (
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/api/v2/cluster"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/protobuf/ptypes/wrappers"
)

func HealthCheck(healthCheck *mesh_core.HealthCheckResource) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&HealthCheckConfigurer{
			healthCheck: healthCheck,
		})
	})
}

type HealthCheckConfigurer struct {
	healthCheck *mesh_core.HealthCheckResource
}

func (e *HealthCheckConfigurer) Configure(cluster *v2.Cluster) error {
	if e.healthCheck == nil {
		return nil
	}
	if e.healthCheck.HasActiveChecks() {
		activeChecks := e.healthCheck.Spec.Conf.GetActiveChecks()
		cluster.HealthChecks = append(cluster.HealthChecks, &envoy_core.HealthCheck{
			HealthChecker: &envoy_core.HealthCheck_TcpHealthCheck_{
				TcpHealthCheck: &envoy_core.HealthCheck_TcpHealthCheck{},
			},
			Interval:           activeChecks.Interval,
			Timeout:            activeChecks.Timeout,
			UnhealthyThreshold: &wrappers.UInt32Value{Value: activeChecks.UnhealthyThreshold},
			HealthyThreshold:   &wrappers.UInt32Value{Value: activeChecks.HealthyThreshold},
		})
	}
	if e.healthCheck.HasPassiveChecks() {
		passiveChecks := e.healthCheck.Spec.Conf.GetPassiveChecks()
		cluster.OutlierDetection = &envoy_cluster.OutlierDetection{
			Interval:        passiveChecks.PenaltyInterval,
			Consecutive_5Xx: &wrappers.UInt32Value{Value: passiveChecks.UnhealthyThreshold},
		}
	}
	return nil
}
