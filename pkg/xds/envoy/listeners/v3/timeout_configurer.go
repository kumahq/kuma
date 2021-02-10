package v3

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type TimeoutConfigurer struct {
	Conf     *mesh_proto.Timeout_Conf
	Protocol core_mesh.Protocol
}

func (c *TimeoutConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if c.Conf == nil {
		return nil
	}

	switch c.Protocol {
	case core_mesh.ProtocolTCP, core_mesh.ProtocolKafka:
		return UpdateTCPProxy(filterChain, func(proxy *envoy_tcp.TcpProxy) error {
			proxy.IdleTimeout = ptypes.DurationProto(c.Conf.GetTCPIdleTimeout())
			return nil
		})
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
			manager.CommonHttpProtocolOptions = &envoy_config_core_v3.HttpProtocolOptions{
				IdleTimeout: ptypes.DurationProto(0),
			}
			return nil
		})
	case core_mesh.ProtocolGRPC:
		return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
			manager.StreamIdleTimeout = ptypes.DurationProto(c.Conf.GetGRPCStreamIdleTimeout())
			return nil
		})
	default:
		return errors.Errorf("unsupported protocol %s", c.Protocol)
	}
}
