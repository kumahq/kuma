package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_filters_listener_original_dst_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/original_dst/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
)

const OriginalDstName = "envoy.filters.listener.original_dst"

type OriginalDstFilterConfigurer struct{}

var _ ListenerConfigurer = &OriginalDstFilterConfigurer{}

func (c *OriginalDstFilterConfigurer) Configure(l *envoy_listener.Listener) error {
	any, err := proto.MarshalAnyDeterministic(&envoy_extensions_filters_listener_original_dst_v3.OriginalDst{})
	if err != nil {
		return err
	}
	l.ListenerFilters = append(l.ListenerFilters, &envoy_listener.ListenerFilter{
		Name: OriginalDstName,
		ConfigType: &envoy_listener.ListenerFilter_TypedConfig{
			TypedConfig: any,
		},
	})
	return nil
}
