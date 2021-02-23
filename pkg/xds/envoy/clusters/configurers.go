package clusters

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v2 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v2"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v3"
)

func OutlierDetection(circuitBreaker *mesh_core.CircuitBreakerResource) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.OutlierDetectionConfigurer{CircuitBreaker: circuitBreaker})
		config.AddV3(&v3.OutlierDetectionConfigurer{CircuitBreaker: circuitBreaker})
	})
}

func ClientSideMTLS(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata, clientService string, tags []envoy.Tags) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.ClientSideMTLSConfigurer{
			Ctx:           ctx,
			Metadata:      metadata,
			ClientService: clientService,
			Tags:          tags,
		})
		config.AddV3(&v3.ClientSideMTLSConfigurer{
			Ctx:           ctx,
			Metadata:      metadata,
			ClientService: clientService,
			Tags:          tags,
		})
	})
}

// UnknownDestinationClientSideMTLS configures cluster with mTLS for a mesh but without extensive destination verification (only Mesh is verified)
func UnknownDestinationClientSideMTLS(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.ClientSideMTLSConfigurer{
			Ctx:           ctx,
			Metadata:      metadata,
			ClientService: "*",
			Tags:          nil,
		})
		config.AddV3(&v3.ClientSideMTLSConfigurer{
			Ctx:           ctx,
			Metadata:      metadata,
			ClientService: "*",
			Tags:          nil,
		})
	})
}

func ClientSideTLS(endpoints []core_xds.Endpoint) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.ClientSideTLSConfigurer{
			Endpoints: endpoints,
		})
		config.AddV3(&v3.ClientSideTLSConfigurer{
			Endpoints: endpoints,
		})
	})
}

func DNSCluster(name string, address string, port uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.DnsClusterConfigurer{
			Name:    name,
			Address: address,
			Port:    port,
		})
		config.AddV2(&v2.AltStatNameConfigurer{})
		config.AddV2(&v2.TimeoutConfigurer{})
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
		config.AddV2(&v2.EdsClusterConfigurer{
			Name: name,
		})
		config.AddV2(&v2.AltStatNameConfigurer{})
		config.AddV3(&v3.EdsClusterConfigurer{
			Name: name,
		})
		config.AddV3(&v3.AltStatNameConfigurer{})
	})
}

func HealthCheck(healthCheck *mesh_core.HealthCheckResource) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.HealthCheckConfigurer{
			HealthCheck: healthCheck,
		})
		config.AddV3(&v3.HealthCheckConfigurer{
			HealthCheck: healthCheck,
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
func LbSubset(keySets [][]string) ClusterBuilderOptFunc {
	return func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.LbSubsetConfigurer{
			KeySets: keySets,
		})
		config.AddV3(&v3.LbSubsetConfigurer{
			KeySets: keySets,
		})
	}
}

func LB(lb *mesh_proto.TrafficRoute_LoadBalancer) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.LbConfigurer{
			Lb: lb,
		})
		config.AddV3(&v3.LbConfigurer{
			Lb: lb,
		})
	})
}

func Timeout(protocol mesh_core.Protocol, conf *mesh_proto.Timeout_Conf) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.TimeoutConfigurer{
			Protocol: protocol,
			Conf:     conf,
		})
		config.AddV3(&v3.TimeoutConfigurer{
			Protocol: protocol,
			Conf:     conf,
		})
	})
}

func DefaultTimeout() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.TimeoutConfigurer{
			Protocol: mesh_core.ProtocolTCP,
		})
		config.AddV3(&v3.TimeoutConfigurer{
			Protocol: mesh_core.ProtocolTCP,
		})
	})
}

func PassThroughCluster(name string) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.PassThroughClusterConfigurer{
			Name: name,
		})
		config.AddV2(&v2.AltStatNameConfigurer{})
		config.AddV2(&v2.TimeoutConfigurer{})
		config.AddV3(&v3.PassThroughClusterConfigurer{
			Name: name,
		})
		config.AddV3(&v3.AltStatNameConfigurer{})
		config.AddV3(&v3.TimeoutConfigurer{})
	})
}

func StaticCluster(name string, address string, port uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.StaticClusterConfigurer{
			Name:    name,
			Address: address,
			Port:    port,
		})
		config.AddV2(&v2.AltStatNameConfigurer{})
		config.AddV2(&v2.TimeoutConfigurer{})
		config.AddV3(&v3.StaticClusterConfigurer{
			Name:    name,
			Address: address,
			Port:    port,
		})
		config.AddV3(&v3.AltStatNameConfigurer{})
		config.AddV3(&v3.TimeoutConfigurer{})
	})
}

func StrictDNSCluster(name string, endpoints []core_xds.Endpoint) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.StrictDNSClusterConfigurer{
			Name:      name,
			Endpoints: endpoints,
		})
		config.AddV2(&v2.AltStatNameConfigurer{})
		config.AddV3(&v3.StrictDNSClusterConfigurer{
			Name:      name,
			Endpoints: endpoints,
		})
		config.AddV3(&v3.AltStatNameConfigurer{})
	})
}

func UpstreamBindConfig(address string, port uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.UpstreamBindConfigConfigurer{
			Address: address,
			Port:    port,
		})
		config.AddV3(&v3.UpstreamBindConfigConfigurer{
			Address: address,
			Port:    port,
		})
	})
}

func Http2() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.AddV2(&v2.Http2Configurer{})
		config.AddV3(&v3.Http2Configurer{})
	})
}
