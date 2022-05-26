package v3

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type TimeoutConfigurer struct {
	Conf     *mesh_proto.Timeout_Conf
	Protocol core_mesh.Protocol
}

func (c *TimeoutConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	switch c.Protocol {
	case core_mesh.ProtocolUnknown, core_mesh.ProtocolTCP, core_mesh.ProtocolKafka:
		return UpdateTCPProxy(filterChain, func(proxy *envoy_tcp.TcpProxy) error {
			proxy.IdleTimeout = util_proto.Duration(c.Conf.GetTcp().GetIdleTimeout().AsDuration())
			return nil
		})
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
			manager.CommonHttpProtocolOptions = &envoy_config_core_v3.HttpProtocolOptions{
				IdleTimeout: util_proto.Duration(0),
			}
			manager.StreamIdleTimeout = util_proto.Duration(0)
			return nil
		})
	case core_mesh.ProtocolGRPC:
		return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
			manager.StreamIdleTimeout = util_proto.Duration(c.Conf.GetGrpc().GetStreamIdleTimeout().AsDuration())
			return nil
		})
	default:
		return errors.Errorf("unsupported protocol %s", c.Protocol)
	}
}
