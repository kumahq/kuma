package filters

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/pkg/errors"
)

func UpdateRouteAction(envoyRoute *envoy_route.Route, updFunc func(*envoy_route.RouteAction) error) error {
	action, ok := envoyRoute.Action.(*envoy_route.Route_Route)
	if !ok {
		return errors.New("expected envoy_route.Route_Route action")
	}
	return updFunc(action.Route)
}
