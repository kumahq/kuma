package v3

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
)

type CommonRouteConfigurationConfigurer struct {
	Name string
}

func (c CommonRouteConfigurationConfigurer) Configure(routeConfiguration *envoy_route.RouteConfiguration) error {
	routeConfiguration.Name = c.Name
	routeConfiguration.ValidateClusters = &wrappers.BoolValue{
		Value: false,
	}
	return nil
}
