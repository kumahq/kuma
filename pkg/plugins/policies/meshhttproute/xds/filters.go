package xds

import (
	"net/http"
	"strings"

	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func routeFilter(filter api.Filter, route *envoy_route.Route) {
	switch filter.Type {
	case api.RequestHeaderModifierType:
		requestHeaderModifier(*filter.RequestHeaderModifier, route)
	case api.ResponseHeaderModifierType:
		responseHeaderModifier(*filter.ResponseHeaderModifier, route)
	}
}

func headerValues(raw string) []string {
	return strings.Split(raw, ",")
}

func headerModifiers(mod api.HeaderModifier) ([]*envoy_config_core.HeaderValueOption, []string) {
	var options []*envoy_config_core.HeaderValueOption

	for _, set := range mod.Set {
		for i, headerValue := range headerValues(set.Value) {
			replace := &envoy_config_core.HeaderValueOption{
				Append: util_proto.Bool(i > 0),
				Header: &envoy_config_core.HeaderValue{
					Key:   http.CanonicalHeaderKey(string(set.Name)),
					Value: headerValue,
				},
			}
			options = append(options, replace)
		}
	}
	for _, add := range mod.Add {
		for _, headerValue := range headerValues(add.Value) {
			appendOption := &envoy_config_core.HeaderValueOption{
				Append: util_proto.Bool(true),
				Header: &envoy_config_core.HeaderValue{
					Key:   http.CanonicalHeaderKey(string(add.Name)),
					Value: headerValue,
				},
			}
			options = append(options, appendOption)
		}
	}

	return options, mod.Remove
}

func requestHeaderModifier(mod api.HeaderModifier, envoyRoute *envoy_route.Route) {
	options, removes := headerModifiers(mod)

	envoyRoute.RequestHeadersToAdd = append(envoyRoute.RequestHeadersToAdd, options...)
	envoyRoute.RequestHeadersToRemove = append(envoyRoute.RequestHeadersToRemove, removes...)
}

func responseHeaderModifier(mod api.HeaderModifier, envoyRoute *envoy_route.Route) {
	options, removes := headerModifiers(mod)

	envoyRoute.ResponseHeadersToAdd = append(envoyRoute.ResponseHeadersToAdd, options...)
	envoyRoute.ResponseHeadersToRemove = append(envoyRoute.ResponseHeadersToRemove, removes...)
}
