package v3

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

// VirtualHostConfigurer is responsible for configuring a single aspect of the entire Envoy VirtualHost,
// such as Route, CORS, etc.
type VirtualHostConfigurer interface {
	// Configure configures a single aspect on a given Envoy VirtualHost.
	Configure(virtualHost *envoy_route.VirtualHost) error
}

// VirtualHostConfigureFunc adapts a configuration function to the
// VirtualHostConfigurer interface.
type VirtualHostConfigureFunc func(vh *envoy_route.VirtualHost) error

func (f VirtualHostConfigureFunc) Configure(vh *envoy_route.VirtualHost) error {
	if f != nil {
		return f(vh)
	}

	return nil
}

// VirtualHostMustConfigureFunc adapts a configuration function that
// never fails to the VirtualHostConfigurer interface.
type VirtualHostMustConfigureFunc func(vh *envoy_route.VirtualHost)

func (f VirtualHostMustConfigureFunc) Configure(vh *envoy_route.VirtualHost) error {
	if f != nil {
		f(vh)
	}

	return nil
}

// RouteConfigurationConfigurer is responsible for configuring a single aspect of the entire Envoy RouteConfiguration,
// such as VirtualHost, HTTP headers to add or remove, etc.
type RouteConfigurationConfigurer interface {
	// Configure configures a single aspect on a given Envoy RouteConfiguration.
	Configure(routeConfiguration *envoy_route.RouteConfiguration) error
}

// RouteConfigurationConfigureFunc adapts a configuration function to the
// RouteConfigurationConfigurer interface.
type RouteConfigurationConfigureFunc func(rc *envoy_route.RouteConfiguration) error

func (f RouteConfigurationConfigureFunc) Configure(rc *envoy_route.RouteConfiguration) error {
	if f != nil {
		return f(rc)
	}

	return nil
}

// RouteConfigurationMustConfigureFunc adapts a configuration function that
// never fails to the RouteConfigurationConfigurer interface.
type RouteConfigurationMustConfigureFunc func(rc *envoy_route.RouteConfiguration)

func (f RouteConfigurationMustConfigureFunc) Configure(rc *envoy_route.RouteConfiguration) error {
	if f != nil {
		f(rc)
	}

	return nil
}
