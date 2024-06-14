package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_filters_http_inspector_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/http_inspector/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
)

const HttpInspectorName = "envoy.filters.listener.http_inspector"

type HTTPInspectorConfigurer struct{}

var _ ListenerConfigurer = &HTTPInspectorConfigurer{}

func (c *HTTPInspectorConfigurer) Configure(l *envoy_listener.Listener) error {
	any, err := proto.MarshalAnyDeterministic(&envoy_extensions_filters_http_inspector_v3.HttpInspector{})
	if err != nil {
		return err
	}
	l.ListenerFilters = append(l.ListenerFilters, &envoy_listener.ListenerFilter{
		Name: "envoy.filters.listener.http_inspector",
		ConfigType: &envoy_listener.ListenerFilter_TypedConfig{
			TypedConfig: any,
		},
	})
	return nil
}
