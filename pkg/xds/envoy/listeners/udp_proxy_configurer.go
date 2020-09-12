package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_udp_proxy "github.com/envoyproxy/go-control-plane/envoy/config/filter/udp/udp_proxy/v2alpha"

	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

func UDPProxy(statsName string, cluster envoy_common.ClusterSubset) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&UDPProxyConfigurer{
			statsName: statsName,
			cluster:   cluster,
		})
	})
}

type UDPProxyConfigurer struct {
	statsName string
	// Cluster to forward traffic to.
	cluster envoy_common.ClusterSubset
}

func (c *UDPProxyConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	udpProxy := c.udpProxy()

	pbst, err := proto.MarshalAnyDeterministic(udpProxy)
	if err != nil {
		return err
	}

	filterChain.Filters = append(filterChain.Filters, &envoy_listener.Filter{
		Name: "envoy.filters.udp_listener.udp_proxy",
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}

func (c *UDPProxyConfigurer) udpProxy() *envoy_udp_proxy.UdpProxyConfig {
	proxy := envoy_udp_proxy.UdpProxyConfig{
		StatPrefix: util_xds.SanitizeMetric(c.statsName),
	}
	proxy.RouteSpecifier = &envoy_udp_proxy.UdpProxyConfig_Cluster{
		Cluster: c.cluster.ClusterName,
	}
	return &proxy
}
