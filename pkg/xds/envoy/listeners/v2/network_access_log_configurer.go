package v2

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
)

const defaultNetworkAccessLogFormat = `[%START_TIME%] %RESPONSE_FLAGS% %KUMA_MESH% %KUMA_SOURCE_ADDRESS_WITHOUT_PORT%(%KUMA_SOURCE_SERVICE%)->%UPSTREAM_HOST%(%KUMA_DESTINATION_SERVICE%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
` // intentional newline at the end

type NetworkAccessLogConfigurer struct {
	AccessLogConfigurer
}

func (c *NetworkAccessLogConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	accessLog, err := convertLoggingBackend(c.AccessLogConfigurer.Mesh, c.AccessLogConfigurer.TrafficDirection, c.AccessLogConfigurer.SourceService, c.AccessLogConfigurer.DestinationService, c.AccessLogConfigurer.Backend, c.AccessLogConfigurer.Proxy, defaultNetworkAccessLogFormat)
	if err != nil {
		return err
	}

	return UpdateTCPProxy(filterChain, func(tcpProxy *envoy_tcp.TcpProxy) error {
		tcpProxy.AccessLog = append(tcpProxy.AccessLog, accessLog)
		return nil
	})
}
