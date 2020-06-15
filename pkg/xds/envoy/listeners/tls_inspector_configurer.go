package listeners

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	_struct "github.com/golang/protobuf/ptypes/struct"
)

func TLSInspector() ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&TLSInspectorConfigurer{})
	})
}

type TLSInspectorConfigurer struct {
}

func (c *TLSInspectorConfigurer) Configure(l *v2.Listener) error {
	l.ListenerFilters = append(l.ListenerFilters, &envoy_api_v2_listener.ListenerFilter{
		Name: wellknown.TlsInspector,
		ConfigType: &envoy_api_v2_listener.ListenerFilter_Config{
			Config: &_struct.Struct{},
		},
	})
	return nil
}
