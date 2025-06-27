package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

type TcpProxyConfigurer struct {
	StatPrefix string
	// Splits to forward traffic to.
	Splits      []envoy_common.Split
	UseMetadata bool
}

func (c *TcpProxyConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if len(c.Splits) == 0 {
		return nil
	}
	tcpProxy := c.tcpProxy()

	pbst, err := proto.MarshalAnyDeterministic(tcpProxy)
	if err != nil {
		return err
	}

	filterChain.Filters = append(filterChain.Filters, &envoy_listener.Filter{
		Name: "envoy.filters.network.tcp_proxy",
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}

func (c *TcpProxyConfigurer) tcpProxy() *envoy_tcp.TcpProxy {
	proxy := envoy_tcp.TcpProxy{
		StatPrefix: util_xds.SanitizeMetric(c.StatPrefix),
	}

	if len(c.Splits) == 1 {
		proxy.ClusterSpecifier = &envoy_tcp.TcpProxy_Cluster{
			Cluster: c.Splits[0].ClusterName(),
		}
		if c.UseMetadata {
			proxy.MetadataMatch = envoy_metadata.LbMetadata(c.Splits[0].LBMetadata())
		}
		return &proxy
	}

	var weightedClusters []*envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight
	for _, split := range c.Splits {
		weightedCluster := &envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight{
			Name:   split.ClusterName(),
			Weight: split.Weight(),
		}
		if c.UseMetadata {
			weightedCluster.MetadataMatch = envoy_metadata.LbMetadata(split.LBMetadata())
		}
		weightedClusters = append(weightedClusters, weightedCluster)
	}
	proxy.ClusterSpecifier = &envoy_tcp.TcpProxy_WeightedClusters{
		WeightedClusters: &envoy_tcp.TcpProxy_WeightedCluster{
			Clusters: weightedClusters,
		},
	}
	return &proxy
}
