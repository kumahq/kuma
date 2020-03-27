package routes

import (
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

func RequestHeadersToAdd(headers map[string]string) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.Add(&RequestHeadersToAddConfigurer{
			headers: headers,
		})
	})
}

type RequestHeadersToAddConfigurer struct {
	headers map[string]string
}

func (r *RequestHeadersToAddConfigurer) Configure(rc *envoy_api_v2.RouteConfiguration) error {
	for key, value := range r.headers {
		rc.RequestHeadersToAdd = append(rc.RequestHeadersToAdd, &envoy_core.HeaderValueOption{
			Header: &envoy_core.HeaderValue{Key: key, Value: value},
		})
	}
	return nil
}
