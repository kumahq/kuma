package filters

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
)

type ResponseHeaderModifierConfigurer struct {
	headerModifier api.HeaderModifier
}

func NewResponseHeaderModifier(modifier api.HeaderModifier) *ResponseHeaderModifierConfigurer {
	return &ResponseHeaderModifierConfigurer{
		headerModifier: modifier,
	}
}

func (f *ResponseHeaderModifierConfigurer) Configure(envoyRoute *envoy_route.Route) error {
	options, removes := headerModifiers(f.headerModifier)

	envoyRoute.ResponseHeadersToAdd = append(envoyRoute.ResponseHeadersToAdd, options...)
	envoyRoute.ResponseHeadersToRemove = append(envoyRoute.ResponseHeadersToRemove, removes...)

	return nil
}
