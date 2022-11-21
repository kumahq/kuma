package routes

import (
	envoy_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

// VirtualHostBuilderOpt is a configuration option for VirtualHostBuilder.
//
// The goal of VirtualHostBuilderOpt is to facilitate fluent VirtualHostBuilder API.
type VirtualHostBuilderOpt interface {
	// ApplyTo adds VirtualHostConfigurer(s) to the VirtualHostBuilder.
	ApplyTo(config *VirtualHostBuilderConfig)
}

func NewVirtualHostBuilder(apiVersion core_xds.APIVersion) *VirtualHostBuilder {
	return &VirtualHostBuilder{
		apiVersion: apiVersion,
	}
}

// VirtualHostBuilder is responsible for generating an Envoy VirtualHost
// by applying a series of VirtualHostConfigurers.
type VirtualHostBuilder struct {
	apiVersion core_xds.APIVersion
	config     VirtualHostBuilderConfig
}

// Configure configures VirtualHostBuilder by adding individual VirtualHostConfigurers.
func (b *VirtualHostBuilder) Configure(opts ...VirtualHostBuilderOpt) *VirtualHostBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy VirtualHost by applying a series of VirtualHostConfigurers.
func (b *VirtualHostBuilder) Build() (envoy.NamedResource, error) {
	switch b.apiVersion {
	case envoy.APIV3:
		virtualHost := envoy_route_v3.VirtualHost{}
		for _, configurer := range b.config.ConfigurersV3 {
			if err := configurer.Configure(&virtualHost); err != nil {
				return nil, err
			}
		}
		return &virtualHost, nil
	default:
		return nil, errors.New("unknown API")
	}
}

// VirtualHostBuilderConfig holds configuration of a VirtualHostBuilder.
type VirtualHostBuilderConfig struct {
	// A series of VirtualHostConfigurers to apply to Envoy VirtualHost.
	ConfigurersV3 []v3.VirtualHostConfigurer
}

// Add appends a given VirtualHostConfigurer to the end of the chain.
func (c *VirtualHostBuilderConfig) AddV3(configurer v3.VirtualHostConfigurer) {
	c.ConfigurersV3 = append(c.ConfigurersV3, configurer)
}

// VirtualHostBuilderOptFunc is a convenience type adapter.
type VirtualHostBuilderOptFunc func(config *VirtualHostBuilderConfig)

func (f VirtualHostBuilderOptFunc) ApplyTo(config *VirtualHostBuilderConfig) {
	if f != nil {
		f(config)
	}
}

// AddVirtualHostConfigurer production an option that adds the given
// configurer to the virtual host builder.
func AddVirtualHostConfigurer(c v3.VirtualHostConfigurer) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(config *VirtualHostBuilderConfig) {
		config.AddV3(c)
	})
}
