package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
)

const defaultNetworkAccessLogFormat = `[%START_TIME%] %RESPONSE_FLAGS% %KUMA_MESH% %KUMA_SOURCE_ADDRESS_WITHOUT_PORT%(%KUMA_SOURCE_SERVICE%)->%UPSTREAM_HOST%(%KUMA_DESTINATION_SERVICE%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes`

type NetworkAccessLogConfigurer struct {
	AccessLogConfigurer
}

func (c *NetworkAccessLogConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	accessLog, err := convertLoggingBackend(c.Mesh, c.TrafficDirection, c.SourceService, c.DestinationService, c.Backend, c.Proxy, defaultNetworkAccessLogFormat)
	if err != nil {
		return err
	}

	return UpdateTCPProxy(filterChain, func(tcpProxy *envoy_tcp.TcpProxy) error {
		tcpProxy.AccessLog = append(tcpProxy.AccessLog, accessLog)
		return nil
	})
}
