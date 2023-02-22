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
	StatsName string
	// Clusters to forward traffic to.
	Clusters    []envoy_common.Cluster
	UseMetadata bool
}

func (c *TcpProxyConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if len(c.Clusters) == 0 {
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
		StatPrefix: util_xds.SanitizeMetric(c.StatsName),
	}

	if len(c.Clusters) == 1 {
		proxy.ClusterSpecifier = &envoy_tcp.TcpProxy_Cluster{
			Cluster: c.Clusters[0].Name(),
		}
		if c.UseMetadata {
			proxy.MetadataMatch = envoy_metadata.LbMetadata(c.Clusters[0].Tags())
		}
		return &proxy
	}

	var weightedClusters []*envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight
	for _, cl := range c.Clusters {
		cluster := cl.(*envoy_common.ClusterImpl)
		weightedCluster := &envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight{
			Name:   cluster.Name(),
			Weight: cluster.Weight(),
		}
		if c.UseMetadata {
			weightedCluster.MetadataMatch = envoy_metadata.LbMetadata(cluster.Tags())
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
