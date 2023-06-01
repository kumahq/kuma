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
	ApplyTo(builder *VirtualHostBuilder)
}

func NewVirtualHostBuilder(apiVersion core_xds.APIVersion) *VirtualHostBuilder {
	return &VirtualHostBuilder{
		apiVersion: apiVersion,
	}
}

// VirtualHostBuilder is responsible for generating an Envoy VirtualHost
// by applying a series of VirtualHostConfigurers.
type VirtualHostBuilder struct {
	apiVersion  core_xds.APIVersion
	configurers []v3.VirtualHostConfigurer
}

// Configure configures VirtualHostBuilder by adding individual VirtualHostConfigurers.
func (b *VirtualHostBuilder) Configure(opts ...VirtualHostBuilderOpt) *VirtualHostBuilder {
	for _, opt := range opts {
		opt.ApplyTo(b)
	}

	return b
}

// Build generates an Envoy VirtualHost by applying a series of VirtualHostConfigurers.
func (b *VirtualHostBuilder) Build() (envoy.NamedResource, error) {
	switch b.apiVersion {
	case envoy.APIV3:
		virtualHost := envoy_route_v3.VirtualHost{}
		for _, configurer := range b.configurers {
			if err := configurer.Configure(&virtualHost); err != nil {
				return nil, err
			}
		}
		return &virtualHost, nil
	default:
		return nil, errors.New("unknown API")
	}
}

// AddConfigurer appends a given VirtualHostConfigurer to the end of the chain.
func (b *VirtualHostBuilder) AddConfigurer(configurer v3.VirtualHostConfigurer) {
	b.configurers = append(b.configurers, configurer)
}

// VirtualHostBuilderOptFunc is a convenience type adapter.
type VirtualHostBuilderOptFunc func(builder *VirtualHostBuilder)

func (f VirtualHostBuilderOptFunc) ApplyTo(builder *VirtualHostBuilder) {
	if f != nil {
		f(builder)
	}
}

// AddVirtualHostConfigurer production an option that adds the given
// configurer to the virtual host builder.
func AddVirtualHostConfigurer(c v3.VirtualHostConfigurer) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(builder *VirtualHostBuilder) {
		builder.AddConfigurer(c)
	})
}
