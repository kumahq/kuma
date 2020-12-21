package v2

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	"github.com/golang/protobuf/ptypes/wrappers"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type MaxConnectAttemptsConfigurer struct {
	Retry *core_mesh.RetryResource
}

func (c *MaxConnectAttemptsConfigurer) Configure(
	filterChain *envoy_listener.FilterChain,
) error {
	return UpdateTCPProxy(filterChain, func(proxy *envoy_tcp.TcpProxy) error {
		proxy.MaxConnectAttempts = &wrappers.UInt32Value{
			Value: c.Retry.Spec.Conf.GetTcp().MaxConnectAttempts,
		}

		return nil
	})
}
