package routes

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func CommonRouteConfiguration(name string) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.Add(&CommonRouteConfigurationConfigurer{
			name: name,
		})
	})
}

type CommonRouteConfigurationConfigurer struct {
	name string
}

func (c CommonRouteConfigurationConfigurer) Configure(routeConfiguration *v2.RouteConfiguration) error {
	routeConfiguration.Name = c.name
	routeConfiguration.ValidateClusters = &wrappers.BoolValue{
		Value: true,
	}
	return nil
}
