package v1alpha1

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

func buildFilterChain(
	proxy *core_xds.Proxy,
	serviceName string,
	splits []envoy_common.Split,
) envoy_listeners.ListenerBuilderOpt {
	tcpProxy := envoy_listeners.TCPProxy(serviceName, splits...)
	builder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, "").
		Configure(tcpProxy)

	return envoy_listeners.FilterChain(builder)
}
