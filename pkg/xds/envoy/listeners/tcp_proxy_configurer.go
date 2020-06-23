package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"github.com/Kong/kuma/pkg/util/proto"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
)

func TcpProxy(statsName string, clusters ...envoy_common.ClusterSubset) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&TcpProxyConfigurer{
			statsName: statsName,
			clusters:  clusters,
		})
	})
}

type TcpProxyConfigurer struct {
	statsName string
	// Clusters to forward traffic to.
	clusters []envoy_common.ClusterSubset
}

func (c *TcpProxyConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	tcpProxy := c.tcpProxy()

	pbst, err := proto.MarshalAnyDeterministic(tcpProxy)
	if err != nil {
		return err
	}

	filterChain.Filters = append(filterChain.Filters, &envoy_listener.Filter{
		Name: envoy_wellknown.TCPProxy,
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}

func (c *TcpProxyConfigurer) tcpProxy() *envoy_tcp.TcpProxy {
	proxy := envoy_tcp.TcpProxy{
		StatPrefix: util_xds.SanitizeMetric(c.statsName),
	}
	if len(c.clusters) == 1 {
		proxy.ClusterSpecifier = &envoy_tcp.TcpProxy_Cluster{
			Cluster: c.clusters[0].ClusterName,
		}
		proxy.MetadataMatch = envoy_common.LbMetadata(c.clusters[0].Tags)
	} else {
		var weightedClusters []*envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight
		for _, cluster := range c.clusters {
			weightedClusters = append(weightedClusters, &envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight{
				Name:          cluster.ClusterName,
				Weight:        cluster.Weight,
				MetadataMatch: envoy_common.LbMetadata(cluster.Tags),
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
