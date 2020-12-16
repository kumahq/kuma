package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/ptypes/wrappers"
)

type CommonRouteConfigurationConfigurer struct {
	Name string
}

func (c CommonRouteConfigurationConfigurer) Configure(routeConfiguration *envoy_api.RouteConfiguration) error {
	routeConfiguration.Name = c.Name
	routeConfiguration.ValidateClusters = &wrappers.BoolValue{
		Value: false,
	}
	return nil
}
