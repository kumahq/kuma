package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/kumahq/kuma/pkg/util/proto"
)

type TLSInspectorConfigurer struct {
}

var _ ListenerConfigurer = &TLSInspectorConfigurer{}

func (c *TLSInspectorConfigurer) Configure(l *envoy_api.Listener) error {
	any, err := proto.MarshalAnyDeterministic(&empty.Empty{})
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
