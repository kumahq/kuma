package listeners

import (
	envoy_api_v2_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/kumahq/kuma/pkg/util/proto"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

func TransparentProxying(transparentProxying *mesh_proto.Dataplane_Networking_TransparentProxying) ListenerBuilderOpt {
	virtual := transparentProxying.GetRedirectPortOutbound() != 0 && transparentProxying.GetRedirectPortInbound() != 0
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		if virtual {
			config.Add(&TransparentProxyingConfigurer{})
		}
	})
}

type TransparentProxyingConfigurer struct {
}

func (c *TransparentProxyingConfigurer) Configure(l *v2.Listener) error {
	any, err := proto.MarshalAnyDeterministic(&empty.Empty{})
	if err != nil {
		return err
	}
	l.Transparent = &wrappers.BoolValue{
		Value: true,
	}
	l.ListenerFilters = append(l.ListenerFilters, &envoy_api_v2_listener.ListenerFilter{
		Name: "envoy.filters.listener.original_dst",
		ConfigType: &envoy_api_v2_listener.ListenerFilter_TypedConfig{
			TypedConfig: any,
		},
	})
	return nil
}
