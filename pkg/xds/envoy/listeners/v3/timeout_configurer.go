package v3

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type TimeoutConfigurer struct {
	Conf     *mesh_proto.Timeout_Conf
	Protocol core_meta.Protocol
}

func (c *TimeoutConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	switch c.Protocol {
	case core_meta.ProtocolUnknown, core_meta.ProtocolTCP, core_meta.ProtocolKafka:
		return UpdateTCPProxy(filterChain, func(proxy *envoy_tcp.TcpProxy) error {
			proxy.IdleTimeout = util_proto.Duration(c.Conf.GetTcp().GetIdleTimeout().AsDuration())
			return nil
		})
	case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
		return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
			c.setIdleTimeout(manager)
			c.setStreamIdleTimeout(manager)
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
	manager.CommonHttpProtocolOptions.IdleTimeout = util_proto.Duration(c.Conf.GetHttp().GetIdleTimeout().AsDuration())
}

func (c *TimeoutConfigurer) setStreamIdleTimeout(manager *envoy_hcm.HttpConnectionManager) {
	// backwards compatibility
	if c.Protocol == core_meta.ProtocolGRPC {
		if sit := c.Conf.GetHttp().GetStreamIdleTimeout(); sit != nil {
			manager.StreamIdleTimeout = sit
		} else {
			manager.StreamIdleTimeout = util_proto.Duration(c.Conf.GetGrpc().GetStreamIdleTimeout().AsDuration())
		}
		return
	}

	manager.StreamIdleTimeout = util_proto.Duration(c.Conf.GetHttp().GetStreamIdleTimeout().AsDuration())
}
