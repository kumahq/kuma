package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_filters_http_inspector_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/http_inspector/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
)

const HttpInspectorName = "envoy.filters.listener.http_inspector"

type HTTPInspectorConfigurer struct {
	DisabledPorts []uint32
}

var _ ListenerConfigurer = &HTTPInspectorConfigurer{}

func (c *HTTPInspectorConfigurer) Configure(l *envoy_listener.Listener) error {
	typedConfig, err := proto.MarshalAnyDeterministic(&envoy_extensions_filters_http_inspector_v3.HttpInspector{})
	if err != nil {
		return err
	}
	listenerFilter := &envoy_listener.ListenerFilter{
		Name: "envoy.filters.listener.http_inspector",
		ConfigType: &envoy_listener.ListenerFilter_TypedConfig{
			TypedConfig: typedConfig,
		},
	}
	if len(c.DisabledPorts) > 0 {
		c.configureFilterDisabled(listenerFilter)
	}
	l.ListenerFilters = append(l.ListenerFilters, listenerFilter)
	return nil
}

func (c *HTTPInspectorConfigurer) configureFilterDisabled(listenerFilter *envoy_listener.ListenerFilter) {
	filterDisabled := &envoy_listener.ListenerFilterChainMatchPredicate{}
	if len(c.DisabledPorts) == 1 {
		filterDisabled.Rule = &envoy_listener.ListenerFilterChainMatchPredicate_DestinationPortRange{
			DestinationPortRange: &typev3.Int32Range{
				Start: int32(c.DisabledPorts[0]),
				End:   int32(c.DisabledPorts[0] + 1),
			},
		}
	} else {
		rules := &envoy_listener.ListenerFilterChainMatchPredicate_OrMatch{
			OrMatch: &envoy_listener.ListenerFilterChainMatchPredicate_MatchSet{
				Rules: []*envoy_listener.ListenerFilterChainMatchPredicate{},
			},
		}
		for _, port := range c.DisabledPorts {
			rules.OrMatch.Rules = append(rules.OrMatch.Rules, &envoy_listener.ListenerFilterChainMatchPredicate{
				Rule: &envoy_listener.ListenerFilterChainMatchPredicate_DestinationPortRange{
					DestinationPortRange: &typev3.Int32Range{
						Start: int32(port),
						End:   int32(port + 1),
					},
				},
			})
		}
		filterDisabled.Rule = rules
	}
	listenerFilter.FilterDisabled = filterDisabled
}
