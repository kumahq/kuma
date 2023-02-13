package listeners

import (
	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

// FilterChainBuilderOpt is a configuration option for FilterChainBuilder.
//
// The goal of FilterChainBuilderOpt is to facilitate fluent FilterChainBuilder API.
type FilterChainBuilderOpt interface {
	// ApplyTo adds FilterChainConfigurer(s) to the FilterChainBuilder.
	ApplyTo(config *FilterChainBuilderConfig)
}

func NewFilterChainBuilder(apiVersion envoy.APIVersion) *FilterChainBuilder {
	return &FilterChainBuilder{
		apiVersion: apiVersion,
	}
}

// FilterChainBuilder is responsible for generating an Envoy filter chain
// by applying a series of FilterChainConfigurers.
type FilterChainBuilder struct {
	apiVersion envoy.APIVersion
	config     FilterChainBuilderConfig
}

// Configure configures FilterChainBuilder by adding individual FilterChainConfigurers.
func (b *FilterChainBuilder) Configure(opts ...FilterChainBuilderOpt) *FilterChainBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy filter chain by applying a series of FilterChainConfigurers.
func (b *FilterChainBuilder) Build() (envoy_types.Resource, error) {
	switch b.apiVersion {
	case envoy.APIV3:
		filterChain := envoy_listener_v3.FilterChain{}

		for _, configurer := range b.config.ConfigurersV3 {
			if err := configurer.Configure(&filterChain); err != nil {
				return nil, err
			}
		}

		// Ensure there is always an HTTP router terminating the filter chain.
		_ = v3.UpdateHTTPConnectionManager(&filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
			router := &envoy_hcm.HttpFilter{
				Name: "envoy.filters.http.router",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: &anypb.Any{
						TypeUrl: "type.googleapis.com/envoy.extensions.filters.http.router.v3.Router",
					},
				},
			}
			hcm.HttpFilters = append(hcm.HttpFilters, router)
			return nil
		})

		return &filterChain, nil

	default:
		return nil, errors.New("unknown API")
	}
}

// FilterChainBuilderConfig holds configuration of a FilterChainBuilder.
type FilterChainBuilderConfig struct {
	// A series of FilterChainConfigurers to apply to Envoy filter chain.
	ConfigurersV3 []v3.FilterChainConfigurer
}

// AddV3 appends a given FilterChainConfigurer to the end of the chain.
func (c *FilterChainBuilderConfig) AddV3(configurer v3.FilterChainConfigurer) {
	c.ConfigurersV3 = append(c.ConfigurersV3, configurer)
}

// FilterChainBuilderOptFunc is a convenience type adapter.
type FilterChainBuilderOptFunc func(config *FilterChainBuilderConfig)

func (f FilterChainBuilderOptFunc) ApplyTo(config *FilterChainBuilderConfig) {
	if f != nil {
		f(config)
	}
}

// AddFilterChainConfigurer produces an option that applies the given
// configurer to the filter chain.
func AddFilterChainConfigurer(c v3.FilterChainConfigurer) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(c)
	})
}
