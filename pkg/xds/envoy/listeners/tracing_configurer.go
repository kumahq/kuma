package listeners

import (
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/proto"
)

func Tracing(backend *v1alpha1.TracingBackend) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&TracingConfigurer{
			backend: backend,
		})
	})
}

type TracingConfigurer struct {
	backend *v1alpha1.TracingBackend
}

func (c *TracingConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if c.backend == nil {
		return nil
	}

	return UpdateFilterConfig(filterChain, envoy_wellknown.HTTPConnectionManager, func(filterConfig proto.Message) error {
		hcm, ok := filterConfig.(*envoy_hcm.HttpConnectionManager)
		if !ok {
			return NewUnexpectedFilterConfigTypeError(filterConfig, &envoy_hcm.HttpConnectionManager{})
		}

		hcm.Tracing = &envoy_hcm.HttpConnectionManager_Tracing{}
		if c.backend.Sampling != 0.0 {
			hcm.Tracing.OverallSampling = &envoy_type.Percent{
				Value: c.backend.Sampling,
			}
		}
		return nil
	})
}
