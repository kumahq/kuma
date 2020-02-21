package listeners

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
)

func Tracing(backend *mesh_proto.TracingBackend) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&TracingConfigurer{
			backend: backend,
		})
	})
}

type TracingConfigurer struct {
	backend *mesh_proto.TracingBackend
}

func (c *TracingConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if c.backend == nil {
		return nil
	}

	return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.Tracing = &envoy_hcm.HttpConnectionManager_Tracing{}
		if c.backend.Sampling != nil {
			hcm.Tracing.OverallSampling = &envoy_type.Percent{
				Value: c.backend.Sampling.Value,
			}
		}
		return nil
	})
}
