package v3

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

// RouteConfigurationConfigurer is responsible for configuring a single aspect of the entire Envoy RouteConfiguration,
// such as VirtualHost, HTTP headers to add or remove, etc.
type RouteConfigurationConfigurer interface {
	// Configure configures a single aspect on a given Envoy RouteConfiguration.
	Configure(routeConfiguration *envoy_config_route_v3.RouteConfiguration) error
}

// RouteConfigurationConfigureFunc adapts a configuration function to the
// RouteConfigurationConfigurer interface.
type RouteConfigurationConfigureFunc func(rc *envoy_config_route_v3.RouteConfiguration) error

func (f RouteConfigurationConfigureFunc) Configure(rc *envoy_config_route_v3.RouteConfiguration) error {
	if f != nil {
		return f(rc)
	}

	return nil
}

// RouteConfigurationMustConfigureFunc adapts a configuration function that
// never fails to the RouteConfigurationConfigurer interface.
type RouteConfigurationMustConfigureFunc func(rc *envoy_config_route_v3.RouteConfiguration)

func (f RouteConfigurationMustConfigureFunc) Configure(rc *envoy_config_route_v3.RouteConfiguration) error {
	if f != nil {
		f(rc)
	}

	return nil
}
