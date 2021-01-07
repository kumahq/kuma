package v2

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_udp_proxy "github.com/envoyproxy/go-control-plane/envoy/config/filter/udp/udp_proxy/v2alpha"

	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type UDPProxyConfigurer struct {
	StatsName string
	// Cluster to forward traffic to.
	Cluster envoy_common.ClusterSubset
}

func (c *UDPProxyConfigurer) Configure(l *v2.Listener) error {
	udpProxy := c.udpProxy()

	pbst, err := proto.MarshalAnyDeterministic(udpProxy)
	if err != nil {
		return err
	}

	l.ListenerFilters = append(l.ListenerFilters, &envoy_api_v2_listener.ListenerFilter{
		Name: "envoy.filters.udp_listener.udp_proxy",
		ConfigType: &envoy_api_v2_listener.ListenerFilter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}

func (c *UDPProxyConfigurer) udpProxy() *envoy_udp_proxy.UdpProxyConfig {
	proxy := envoy_udp_proxy.UdpProxyConfig{
		StatPrefix: util_xds.SanitizeMetric(c.StatsName),
	}
	proxy.RouteSpecifier = &envoy_udp_proxy.UdpProxyConfig_Cluster{
		Cluster: c.Cluster.ClusterName,
	}
	return &proxy
}
