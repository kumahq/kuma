package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	"github.com/golang/protobuf/ptypes/wrappers"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

func MaxConnectAttempts(retry *core_mesh.RetryResource) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		if retry != nil && retry.Spec.Conf.GetTcp() != nil {
			config.Add(&MaxConnectAttemptsConfigurer{
				retry: retry,
			})
		}
	})
}

type MaxConnectAttemptsConfigurer struct {
	retry *core_mesh.RetryResource
}

func (c *MaxConnectAttemptsConfigurer) Configure(
	filterChain *envoy_listener.FilterChain,
) error {
	return UpdateTCPProxy(filterChain, func(proxy *envoy_tcp.TcpProxy) error {
		proxy.MaxConnectAttempts = &wrappers.UInt32Value{
			Value: c.retry.Spec.Conf.GetTcp().MaxConnectAttempts,
		}

		return nil
	})
}
