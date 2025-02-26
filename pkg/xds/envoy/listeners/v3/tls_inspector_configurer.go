package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_filters_listener_tls_inspector_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
)

const TlsInspectorName = "envoy.filters.listener.tls_inspector"

type TLSInspectorConfigurer struct {
	DisabledPorts []uint32
}

var _ ListenerConfigurer = &TLSInspectorConfigurer{}

func (c *TLSInspectorConfigurer) Configure(l *envoy_listener.Listener) error {
	any, err := proto.MarshalAnyDeterministic(&envoy_extensions_filters_listener_tls_inspector_v3.TlsInspector{})
	if err != nil {
		return err
	}
	listenerFilter := &envoy_listener.ListenerFilter{
		Name: TlsInspectorName,
		ConfigType: &envoy_listener.ListenerFilter_TypedConfig{
			TypedConfig: any,
		},
	}
	if len(c.DisabledPorts) > 0 {
		c.configureFilterDisabled(listenerFilter)
	}
	l.ListenerFilters = append(l.ListenerFilters, listenerFilter)
	return nil
}

func (c *TLSInspectorConfigurer) configureFilterDisabled(listenerFilter *envoy_listener.ListenerFilter) {
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
