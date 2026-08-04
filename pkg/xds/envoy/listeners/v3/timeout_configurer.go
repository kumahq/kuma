package v3

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/pkg/errors"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

type TimeoutConfigurer struct {
	Conf     envoy_common.Timeouts
	Protocol core_meta.Protocol
}

func (c *TimeoutConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	switch c.Protocol {
	case core_meta.ProtocolUnknown, core_meta.ProtocolTCP, core_meta.ProtocolKafka:
		return UpdateTCPProxy(filterChain, func(proxy *envoy_tcp.TcpProxy) error {
			proxy.IdleTimeout = util_proto.Duration(c.Conf.TcpIdle)
			return nil
		})
	case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
		return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
			c.setIdleTimeout(manager)
			manager.StreamIdleTimeout = util_proto.Duration(c.Conf.HttpStreamIdle)
			return nil
		})
	default:
		return errors.Errorf("unsupported protocol %s", c.Protocol)
	}
}

func (c *TimeoutConfigurer) setIdleTimeout(manager *envoy_hcm.HttpConnectionManager) {
	if manager.CommonHttpProtocolOptions == nil {
		manager.CommonHttpProtocolOptions = &envoy_config_core_v3.HttpProtocolOptions{}
	}
	manager.CommonHttpProtocolOptions.IdleTimeout = util_proto.Duration(c.Conf.HttpIdle)
}
