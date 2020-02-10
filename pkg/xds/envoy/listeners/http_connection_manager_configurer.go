package listeners

import (
	"github.com/golang/protobuf/ptypes"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	util_error "github.com/Kong/kuma/pkg/util/error"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func HttpConnectionManager(statsName string) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&HttpConnectionManagerConfigurer{
			statsName: statsName,
		})
	})
}

type HttpConnectionManagerConfigurer struct {
	statsName string
}

func (c *HttpConnectionManagerConfigurer) Configure(l *v2.Listener) error {
	config := &envoy_hcm.HttpConnectionManager{
		StatPrefix: util_xds.SanitizeMetric(c.statsName),
		CodecType:  envoy_hcm.HttpConnectionManager_AUTO,
		HttpFilters: []*envoy_hcm.HttpFilter{
			{Name: envoy_wellknown.Router},
		},
		// notice that route configuration is left up to other configurers
	}

	pbst, err := ptypes.MarshalAny(config)
	util_error.MustNot(err)

	for i := range l.FilterChains {
		l.FilterChains[i].Filters = append(l.FilterChains[i].Filters, &envoy_listener.Filter{
			Name: envoy_wellknown.HTTPConnectionManager,
			ConfigType: &envoy_listener.Filter_TypedConfig{
				TypedConfig: pbst,
			},
		})
	}

	return nil
}
