package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_filters_listener_tls_inspector_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
)

type TLSInspectorConfigurer struct{}

var _ ListenerConfigurer = &TLSInspectorConfigurer{}

func (c *TLSInspectorConfigurer) Configure(l *envoy_listener.Listener) error {
	any, err := proto.MarshalAnyDeterministic(&envoy_extensions_filters_listener_tls_inspector_v3.TlsInspector{})
	if err != nil {
		return err
	}
	l.ListenerFilters = append(l.ListenerFilters, &envoy_listener.ListenerFilter{
		Name: "envoy.filters.listener.tls_inspector",
		ConfigType: &envoy_listener.ListenerFilter_TypedConfig{
			TypedConfig: any,
		},
	})
	return nil
}
