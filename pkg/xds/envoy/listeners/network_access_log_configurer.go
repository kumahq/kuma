package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

const defaultNetworkAccessLogFormat = `[%START_TIME%] %RESPONSE_FLAGS% %KUMA_MESH% %KUMA_SOURCE_ADDRESS_WITHOUT_PORT%(%KUMA_SOURCE_SERVICE%)->%UPSTREAM_HOST%(%KUMA_DESTINATION_SERVICE%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
` // intentional newline at the end

func NetworkAccessLog(mesh string, trafficDirection string, sourceService string, destinationService string, backend *v1alpha1.LoggingBackend, proxy *core_xds.Proxy) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		if backend != nil {
			config.Add(&NetworkAccessLogConfigurer{
				AccessLogConfigurer: AccessLogConfigurer{
					mesh:               mesh,
					trafficDirection:   trafficDirection,
					sourceService:      sourceService,
					destinationService: destinationService,
					backend:            backend,
					proxy:              proxy,
				},
			})
		}
	})
}

type NetworkAccessLogConfigurer struct {
	AccessLogConfigurer
}

func (c *NetworkAccessLogConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	accessLog, err := convertLoggingBackend(c.AccessLogConfigurer.mesh, c.AccessLogConfigurer.trafficDirection, c.AccessLogConfigurer.sourceService, c.AccessLogConfigurer.destinationService, c.AccessLogConfigurer.backend, c.AccessLogConfigurer.proxy, defaultNetworkAccessLogFormat)
	if err != nil {
		return err
	}

	return UpdateTCPProxy(filterChain, func(tcpProxy *envoy_tcp.TcpProxy) error {
		tcpProxy.AccessLog = append(tcpProxy.AccessLog, accessLog)
		return nil
	})
}
