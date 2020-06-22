package listeners

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/Kong/kuma/pkg/util/proto"
)

func TLSInspector() ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&TLSInspectorConfigurer{})
	})
}

type TLSInspectorConfigurer struct {
}

func (c *TLSInspectorConfigurer) Configure(l *v2.Listener) error {
	any, err := proto.MarshalAnyDeterministic(&empty.Empty{})
	if err != nil {
		return err
	}
	l.ListenerFilters = append(l.ListenerFilters, &envoy_api_v2_listener.ListenerFilter{
		Name: "envoy.filters.listener.tls_inspector",
		ConfigType: &envoy_api_v2_listener.ListenerFilter_TypedConfig{
			TypedConfig: any,
		},
	})
	return nil
}
