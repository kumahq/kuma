package clusters

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v3"
	endpoints_v3 "github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
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

func ClientSideMTLS(ctx xds_context.Context, upstreamService string, tags []envoy.Tags) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.ClientSideMTLSConfigurer{
			Ctx:             ctx,
			UpstreamService: upstreamService,
			Tags:            tags,
		})
	})
}

// UnknownDestinationClientSideMTLS configures cluster with mTLS for a mesh but without extensive destination verification (only Mesh is verified)
func UnknownDestinationClientSideMTLS(ctx xds_context.Context) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.ClientSideMTLSConfigurer{
			Ctx:             ctx,
			UpstreamService: "*",
			Tags:            nil,
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

func DNSCluster(name string, address string, port uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.DnsClusterConfigurer{
			Name:    name,
			Address: address,
			Port:    port,
		})
		config.AddV3(&v3.AltStatNameConfigurer{})
		config.AddV3(&v3.TimeoutConfigurer{})
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
// 1) TrafficRoute for splitting traffic. Example: TrafficRoute that splits 10% of the traffic to version 1 of the service backend and 90% traffic to version 2 of the service backend
// 2) Multiple outbound sections with the same service
//    Example:
//    type: Dataplane
//    networking:
//      outbound:
//      - port: 1234
//        tags:
//          kuma.io/service: backend
//      - port: 1234
//        tags:
//          kuma.io/service: backend
//          version: v1
//    Only one cluster "backend" is generated for such dataplane, but with lb subset by version.
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

func Timeout(protocol core_mesh.Protocol, conf *mesh_proto.Timeout_Conf) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.TimeoutConfigurer{
			Protocol: protocol,
			Conf:     conf,
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
		config.AddV3(&v3.TimeoutConfigurer{})
	})
}

func StaticCluster(name string, address string, port uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.StaticClusterConfigurer{
			Name:           name,
			LoadAssignment: endpoints_v3.CreateStaticEndpoint(name, address, port),
		})
		config.AddV3(&v3.AltStatNameConfigurer{})
		config.AddV3(&v3.TimeoutConfigurer{})
	})
}

func StaticClusterUnixSocket(name string, path string) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.StaticClusterConfigurer{
			Name:           name,
			LoadAssignment: endpoints_v3.CreateStaticEndpointUnixSocket(name, path),
		})
		config.AddV3(&v3.AltStatNameConfigurer{})
		config.AddV3(&v3.TimeoutConfigurer{})
	})
}

func StrictDNSCluster(name string, endpoints []core_xds.Endpoint, hasIPv6 bool) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.StrictDNSClusterConfigurer{
			Name:      name,
			Endpoints: endpoints,
			HasIPv6:   hasIPv6,
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

func Http2() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.Http2Configurer{})
	})
}

func Http() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV3(&v3.HttpConfigurer{})
	})
}
