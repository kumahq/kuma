package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
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
