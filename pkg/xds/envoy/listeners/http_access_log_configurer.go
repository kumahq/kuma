package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

const defaultHttpAccessLogFormat = `[%START_TIME%] %KUMA_MESH% "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%KUMA_SOURCE_SERVICE%" "%KUMA_DESTINATION_SERVICE%" "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%" "%UPSTREAM_HOST%"
` // intentional newline at the end

func HttpAccessLog(mesh string, trafficDirection string, sourceService string, destinationService string, backend *mesh_proto.LoggingBackend, proxy *core_xds.Proxy) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		if backend != nil {
			config.Add(&HttpAccessLogConfigurer{
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

type HttpAccessLogConfigurer struct {
	AccessLogConfigurer
}

func (c *HttpAccessLogConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	accessLog, err := convertLoggingBackend(c.AccessLogConfigurer.mesh, c.AccessLogConfigurer.trafficDirection, c.AccessLogConfigurer.sourceService, c.AccessLogConfigurer.destinationService, c.AccessLogConfigurer.backend, c.AccessLogConfigurer.proxy, defaultHttpAccessLogFormat)
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.AccessLog = append(hcm.AccessLog, accessLog)
		return nil
	})
}
