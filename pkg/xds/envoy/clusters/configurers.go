package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v3"
)

func OutlierDetection(circuitBreaker *core_mesh.CircuitBreakerResource) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.OutlierDetectionConfigurer{CircuitBreaker: circuitBreaker})
	})
}

func CircuitBreaker(circuitBreaker *core_mesh.CircuitBreakerResource) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.CircuitBreakerConfigurer{CircuitBreaker: circuitBreaker})
	})
}

func ClientSideMTLS(tracker core_xds.SecretsTracker, mesh *core_mesh.MeshResource, upstreamService string, upstreamTLSReady bool, tags []envoy.Tags) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.ClientSideMTLSConfigurer{
			SecretsTracker:   tracker,
			UpstreamMesh:     mesh,
			UpstreamService:  upstreamService,
			LocalMesh:        mesh,
			Tags:             tags,
			UpstreamTLSReady: upstreamTLSReady,
		})
	})
}

func CrossMeshClientSideMTLS(tracker core_xds.SecretsTracker, localMesh *core_mesh.MeshResource, upstreamMesh *core_mesh.MeshResource, upstreamService string, upstreamTLSReady bool, tags []envoy.Tags) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.ClientSideMTLSConfigurer{
			SecretsTracker:   tracker,
			UpstreamMesh:     upstreamMesh,
			UpstreamService:  upstreamService,
			LocalMesh:        localMesh,
			Tags:             tags,
			UpstreamTLSReady: upstreamTLSReady,
		})
	})
}

// UnknownDestinationClientSideMTLS configures cluster with mTLS for a mesh but without extensive destination verification (only Mesh is verified)
func UnknownDestinationClientSideMTLS(tracker core_xds.SecretsTracker, mesh *core_mesh.MeshResource) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.ClientSideMTLSConfigurer{
			SecretsTracker:   tracker,
			UpstreamMesh:     mesh,
			UpstreamService:  "*",
			LocalMesh:        mesh,
			Tags:             nil,
			UpstreamTLSReady: true,
		})
	})
}

func ClientSideTLS(endpoints []core_xds.Endpoint) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.ClientSideTLSConfigurer{
			Endpoints: endpoints,
		})
	})
}

func EdsCluster(name string) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.EdsClusterConfigurer{
			Name: name,
		})
		config.AddV3(&v3.AltStatNameConfigurer{})
	})
}

// ProvidedEndpointCluster sets the cluster with the defined endpoints, this is useful when endpoints are not discovered using EDS, so we don't use EdsCluster
func ProvidedEndpointCluster(name string, hasIPv6 bool, endpoints ...core_xds.Endpoint) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.ProvidedEndpointClusterConfigurer{
			Name:      name,
			Endpoints: endpoints,
			HasIPv6:   hasIPv6,
		})
		config.AddV3(&v3.AltStatNameConfigurer{})
	})
}

func HealthCheck(protocol core_mesh.Protocol, healthCheck *core_mesh.HealthCheckResource) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.HealthCheckConfigurer{
			HealthCheck: healthCheck,
			Protocol:    protocol,
		})
	})
}

// LbSubset is required for MetadataMatch in Weighted Cluster in TCP Proxy to work.
// Subset loadbalancing is used in two use cases
//  1. TrafficRoute for splitting traffic. Example: TrafficRoute that splits 10% of the traffic to version 1 of the service backend and 90% traffic to version 2 of the service backend
//  2. Multiple outbound sections with the same service
//     Example:
//     type: Dataplane
//     networking:
//     outbound:
//     - port: 1234
//     tags:
//     kuma.io/service: backend
//     - port: 1234
//     tags:
//     kuma.io/service: backend
//     version: v1
//     Only one cluster "backend" is generated for such dataplane, but with lb subset by version.
func LbSubset(tagSets envoy.TagKeysSlice) ClusterBuilderOptFunc {
	return func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.LbSubsetConfigurer{
			TagKeysSets: tagSets,
		})
	}
}

func LB(lb *mesh_proto.TrafficRoute_LoadBalancer) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.LbConfigurer{
			Lb: lb,
		})
	})
}

func Timeout(timeout *mesh_proto.Timeout_Conf, protocol core_mesh.Protocol) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.TimeoutConfigurer{
			Protocol: protocol,
			Conf:     timeout,
		})
	})
}

func DefaultTimeout() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.TimeoutConfigurer{
			Protocol: core_mesh.ProtocolTCP,
		})
	})
}

func PassThroughCluster(name string) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.PassThroughClusterConfigurer{
			Name: name,
		})
		config.AddV3(&v3.AltStatNameConfigurer{})
	})
}

func UpstreamBindConfig(address string, port uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.UpstreamBindConfigConfigurer{
			Address: address,
			Port:    port,
		})
	})
}

func ConnectionBufferLimit(bytes uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(v3.ClusterMustConfigureFunc(func(c *envoy_cluster.Cluster) {
			c.PerConnectionBufferLimitBytes = wrapperspb.UInt32(bytes)
		}))
	})
}

func Http2() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.Http2Configurer{})
	})
}

func Http2FromEdge() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.Http2Configurer{EdgeProxyWindowSizes: true})
	})
}

func Http() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.HttpConfigurer{})
	})
}
