package clusters

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/api/v2/cluster"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	"github.com/Kong/kuma/pkg/xds/envoy"
	envoy_endpoints "github.com/Kong/kuma/pkg/xds/envoy/endpoints"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
)

const (
	defaultConnectTimeout = 5 * time.Second
)

func CreateLocalCluster(clusterName string, address string, port uint32) *v2.Cluster {
	return clusterWithAltStatName(&v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(defaultConnectTimeout),
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_STATIC},
		LoadAssignment:       envoy_endpoints.CreateStaticEndpoint(clusterName, address, port),
	})
}

func CreateEdsCluster(ctx xds_context.Context, clusterName string, metadata *core_xds.DataplaneMetadata) (*v2.Cluster, error) {
	tlsContext, err := envoy.CreateUpstreamTlsContext(ctx, metadata)
	if err != nil {
		return nil, err
	}
	return clusterWithAltStatName(&v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(defaultConnectTimeout),
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_EDS},
		EdsClusterConfig: &v2.Cluster_EdsClusterConfig{
			EdsConfig: &envoy_core.ConfigSource{
				ConfigSourceSpecifier: &envoy_core.ConfigSource_Ads{
					Ads: &envoy_core.AggregatedConfigSource{},
				},
			},
		},
		TlsContext: tlsContext,
	}), nil
}

func clusterWithAltStatName(cluster *v2.Cluster) *v2.Cluster {
	sanitizedName := util_xds.SanitizeMetric(cluster.Name)
	if sanitizedName != cluster.Name {
		cluster.AltStatName = sanitizedName
	}
	return cluster
}

func ClusterWithHealthChecks(cluster *v2.Cluster, healthCheck *mesh_core.HealthCheckResource) *v2.Cluster {
	if healthCheck == nil {
		return cluster
	}
	if healthCheck.HasActiveChecks() {
		activeChecks := healthCheck.Spec.Conf.GetActiveChecks()
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
	if healthCheck.HasPassiveChecks() {
		passiveChecks := healthCheck.Spec.Conf.GetPassiveChecks()
		cluster.OutlierDetection = &envoy_cluster.OutlierDetection{
			Interval:        passiveChecks.PenaltyInterval,
			Consecutive_5Xx: &wrappers.UInt32Value{Value: passiveChecks.UnhealthyThreshold},
		}
	}
	return cluster
}

func CreatePassThroughCluster(clusterName string) *v2.Cluster {
	return clusterWithAltStatName(&v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(defaultConnectTimeout),
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_ORIGINAL_DST},
		LbPolicy:             v2.Cluster_ORIGINAL_DST_LB,
	})
}
