package routes

import (
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func RequestHeadersToRemove(headers ...string) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.Add(&RequestHeadersToRemoveConfigurer{
			headers: headers,
		})
	})
}

type RequestHeadersToRemoveConfigurer struct {
	headers []string
}

func (r *RequestHeadersToRemoveConfigurer) Configure(rc *envoy_api_v2.RouteConfiguration) error {
	rc.RequestHeadersToRemove = append(rc.RequestHeadersToRemove, r.headers...)
	return nil
}
