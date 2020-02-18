package listeners

import (
	"github.com/golang/protobuf/ptypes"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func HttpConnectionManager(statsName string) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&HttpConnectionManagerConfigurer{
			statsName: statsName,
		})
	})
}

type HttpConnectionManagerConfigurer struct {
	statsName string
}

func (c *HttpConnectionManagerConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	config := &envoy_hcm.HttpConnectionManager{
		StatPrefix: util_xds.SanitizeMetric(c.statsName),
		CodecType:  envoy_hcm.HttpConnectionManager_AUTO,
		HttpFilters: []*envoy_hcm.HttpFilter{
			{Name: envoy_wellknown.Router},
		},
		// notice that route configuration is left up to other configurers
	}

	pbst, err := ptypes.MarshalAny(config)
	if err != nil {
		return err
	}

	filterChain.Filters = append(filterChain.Filters, &envoy_listener.Filter{
		Name: envoy_wellknown.HTTPConnectionManager,
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}
