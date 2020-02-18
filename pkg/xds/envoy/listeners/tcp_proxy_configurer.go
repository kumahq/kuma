package listeners

import (
	"github.com/golang/protobuf/ptypes"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	util_xds "github.com/Kong/kuma/pkg/util/xds"
	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
)

func TcpProxy(statsName string, clusters ...envoy_common.ClusterInfo) FilterChainBuilderOpt {
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
	clusters []envoy_common.ClusterInfo
}

func (c *TcpProxyConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	tcpProxy := c.tcpProxy()

	pbst, err := ptypes.MarshalAny(tcpProxy)
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
			Cluster: c.clusters[0].Name,
		}
	} else {
		var weightedClusters []*envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight
		for _, cluster := range c.clusters {
			weightedClusters = append(weightedClusters, &envoy_tcp.TcpProxy_WeightedCluster_ClusterWeight{
				Name:   cluster.Name,
				Weight: cluster.Weight,
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
