package v2

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"

	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v2"
)

type TcpProxyConfigurer struct {
	StatsName string
	// Clusters to forward traffic to.
	Clusters []envoy_common.ClusterSubset
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
			Cluster: c.Clusters[0].ClusterName,
		}
		proxy.MetadataMatch = envoy_metadata.LbMetadata(c.Clusters[0].Tags)
	} else {
		var weightedClusters []*envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight
		for _, cluster := range c.Clusters {
			weightedClusters = append(weightedClusters, &envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight{
				Name:          cluster.ClusterName,
				Weight:        cluster.Weight,
				MetadataMatch: envoy_metadata.LbMetadata(cluster.Tags),
			})
		}
		proxy.ClusterSpecifier = &envoy_tcp.TcpProxy_WeightedClusters{
			WeightedClusters: &envoy_tcp.TcpProxy_WeightedCluster{
				Clusters: weightedClusters,
			},
		}
	}
	return &proxy
}
