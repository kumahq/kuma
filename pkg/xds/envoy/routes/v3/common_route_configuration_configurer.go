package v3

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type CommonRouteConfigurationConfigurer struct{}

func (c CommonRouteConfigurationConfigurer) Configure(routeConfiguration *envoy_config_route_v3.RouteConfiguration) error {
	routeConfiguration.ValidateClusters = util_proto.Bool(false)
	return nil
}
