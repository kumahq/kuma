package filters

import (
	"strings"

	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type RequestHeaderModifierConfigurer struct {
	headerModifier *api.HeaderModifier
}

func NewRequestHeaderModifier(modifier *api.HeaderModifier) *RequestHeaderModifierConfigurer {
	return &RequestHeaderModifierConfigurer{
		headerModifier: modifier,
	}
}

func (f *RequestHeaderModifierConfigurer) Configure(envoyRoute *envoy_route.Route) error {
	if f.headerModifier == nil {
		return nil
	}

	options, removes := headerModifiers(*f.headerModifier)

	envoyRoute.RequestHeadersToAdd = append(envoyRoute.RequestHeadersToAdd, options...)
	envoyRoute.RequestHeadersToRemove = append(envoyRoute.RequestHeadersToRemove, removes...)

	return nil
}

func headerModifiers(mod api.HeaderModifier) ([]*envoy_config_core.HeaderValueOption, []string) {
	var options []*envoy_config_core.HeaderValueOption

	for _, set := range mod.Set {
		for i, headerValue := range headerValues(set.Value) {
			replace := &envoy_config_core.HeaderValueOption{
				Append: util_proto.Bool(i > 0),
				Header: &envoy_config_core.HeaderValue{
					Key:   string(set.Name),
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
					Key:   string(add.Name),
					Value: headerValue,
				},
			}
			options = append(options, appendOption)
		}
	}

	return options, mod.Remove
}

func headerValues(raw common_api.HeaderValue) []string {
	return strings.Split(string(raw), ",")
}
