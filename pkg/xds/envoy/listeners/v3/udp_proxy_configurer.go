package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_udp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type UDPProxyConfigurer struct {
	StatsName string
	// Cluster to forward traffic to.
	Cluster envoy_common.ClusterSubset
}

func (c *UDPProxyConfigurer) Configure(l *envoy_listener.Listener) error {
	udpProxy := c.udpProxy()

	pbst, err := proto.MarshalAnyDeterministic(udpProxy)
	if err != nil {
		return err
	}

	l.ListenerFilters = append(l.ListenerFilters, &envoy_listener.ListenerFilter{
		Name: "envoy.filters.udp_listener.udp_proxy",
		ConfigType: &envoy_listener.ListenerFilter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}

func (c *UDPProxyConfigurer) udpProxy() *envoy_udp.UdpProxyConfig {
	proxy := envoy_udp.UdpProxyConfig{
		StatPrefix: util_xds.SanitizeMetric(c.StatsName),
	}
	proxy.RouteSpecifier = &envoy_udp.UdpProxyConfig_Cluster{
		Cluster: c.Cluster.ClusterName,
	}
	return &proxy
}
