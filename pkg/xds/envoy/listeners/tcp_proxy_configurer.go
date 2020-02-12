package listeners

import (
	"github.com/golang/protobuf/ptypes"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

type ClusterInfo struct {
	Name   string
	Weight uint32
	Tags   map[string]string
}

func TcpProxy(statsName string, clusters ...ClusterInfo) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&TcpProxyConfigurer{
			statsName: statsName,
			clusters:  clusters,
		})
	})
}

type TcpProxyConfigurer struct {
	statsName string
	// Clusters to forward traffic to.
	clusters []ClusterInfo
}

func (c *TcpProxyConfigurer) Configure(l *v2.Listener) error {
	tcpProxy := c.tcpProxy()

	pbst, err := ptypes.MarshalAny(tcpProxy)
	if err != nil {
		return err
	}

	for i := range l.FilterChains {
		l.FilterChains[i].Filters = append(l.FilterChains[i].Filters, &envoy_listener.Filter{
			Name: envoy_wellknown.TCPProxy,
			ConfigType: &envoy_listener.Filter_TypedConfig{
				TypedConfig: pbst,
			},
		})
	}

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
