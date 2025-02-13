package filters

import (
	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type RequestHeaderModifierConfigurer struct {
	headerModifier api.HeaderModifier
}

func NewRequestHeaderModifier(modifier api.HeaderModifier) *RequestHeaderModifierConfigurer {
	return &RequestHeaderModifierConfigurer{
		headerModifier: modifier,
	}
}

func (f *RequestHeaderModifierConfigurer) Configure(envoyRoute *envoy_route.Route) error {
	options, removes := headerModifiers(f.headerModifier)

	envoyRoute.RequestHeadersToAdd = append(envoyRoute.RequestHeadersToAdd, options...)
	envoyRoute.RequestHeadersToRemove = append(envoyRoute.RequestHeadersToRemove, removes...)

	return nil
}

func headerModifiers(mod api.HeaderModifier) ([]*envoy_config_core.HeaderValueOption, []string) {
	var options []*envoy_config_core.HeaderValueOption

	for _, set := range pointer.Deref(mod.Set) {
		replace := &envoy_config_core.HeaderValueOption{
			AppendAction: envoy_config_core.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			Header: &envoy_config_core.HeaderValue{
				Key:   string(set.Name),
				Value: string(set.Value),
			},
		}
		options = append(options, replace)
	}
	for _, add := range pointer.Deref(mod.Add) {
		appendOption := &envoy_config_core.HeaderValueOption{
			AppendAction: envoy_config_core.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
			Header: &envoy_config_core.HeaderValue{
				Key:   string(add.Name),
				Value: string(add.Value),
			},
		}
		options = append(options, appendOption)
	}

	return options, pointer.Deref(mod.Remove)
}
