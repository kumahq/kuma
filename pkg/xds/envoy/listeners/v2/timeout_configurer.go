package v2

import (
	envoy_api_v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
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
	switch c.Protocol {
	case core_mesh.ProtocolUnknown, core_mesh.ProtocolTCP, core_mesh.ProtocolKafka:
		return UpdateTCPProxy(filterChain, func(proxy *envoy_tcp.TcpProxy) error {
			proxy.IdleTimeout = ptypes.DurationProto(c.Conf.GetTcp().GetIdleTimeout().AsDuration())
			return nil
		})
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
			manager.CommonHttpProtocolOptions = &envoy_api_v2_core.HttpProtocolOptions{
				IdleTimeout: ptypes.DurationProto(0),
			}
			return nil
		})
	case core_mesh.ProtocolGRPC:
		return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
			manager.StreamIdleTimeout = ptypes.DurationProto(c.Conf.GetGrpc().GetStreamIdleTimeout().AsDuration())
			return nil
		})
	default:
		return errors.Errorf("unsupported protocol %s", c.Protocol)
	}
}
