package listeners

import (
	envoy_listener_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/xds/envoy"
	v2 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v2"
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
	case envoy.APIV2:
		filterChain := envoy_listener_v2.FilterChain{}
		for _, configurer := range b.config.ConfigurersV2 {
			if err := configurer.Configure(&filterChain); err != nil {
				return nil, err
			}
		}
		return &filterChain, nil
	case envoy.APIV3:
		filterChain := envoy_listener_v3.FilterChain{}
		for _, configurer := range b.config.ConfigurersV3 {
			if err := configurer.Configure(&filterChain); err != nil {
				return nil, err
			}
		}
		return &filterChain, nil
	default:
		return nil, errors.New("unknown API")
	}
}

// FilterChainBuilderConfig holds configuration of a FilterChainBuilder.
type FilterChainBuilderConfig struct {
	// A series of FilterChainConfigurers to apply to Envoy filter chain.
	ConfigurersV2 []v2.FilterChainConfigurer
	ConfigurersV3 []v3.FilterChainConfigurer
}

// Add appends a given FilterChainConfigurer to the end of the chain.
func (c *FilterChainBuilderConfig) AddV2(configurer v2.FilterChainConfigurer) {
	c.ConfigurersV2 = append(c.ConfigurersV2, configurer)
}

// Add appends a given FilterChainConfigurer to the end of the chain.
func (c *FilterChainBuilderConfig) AddV3(configurer v3.FilterChainConfigurer) {
	c.ConfigurersV3 = append(c.ConfigurersV3, configurer)
}

// FilterChainBuilderOptFunc is a convenience type adapter.
type FilterChainBuilderOptFunc func(config *FilterChainBuilderConfig)

func (f FilterChainBuilderOptFunc) ApplyTo(config *FilterChainBuilderConfig) {
	f(config)
}
