package listeners

import (
	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

// ListenerBuilderOpt is a configuration option for ListenerBuilder.
//
// The goal of ListenerBuilderOpt is to facilitate fluent ListenerBuilder API.
type ListenerBuilderOpt interface {
	// ApplyTo adds ListenerConfigurer(s) to the ListenerBuilder.
	ApplyTo(config *ListenerBuilderConfig)
}

func NewListenerBuilder(apiVersion core_xds.APIVersion) *ListenerBuilder {
	return &ListenerBuilder{
		apiVersion: apiVersion,
	}
}

// ListenerBuilder is responsible for generating an Envoy listener
// by applying a series of ListenerConfigurers.
type ListenerBuilder struct {
	apiVersion core_xds.APIVersion
	config     ListenerBuilderConfig
}

// Configure configures ListenerBuilder by adding individual ListenerConfigurers.
func (b *ListenerBuilder) Configure(opts ...ListenerBuilderOpt) *ListenerBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy listener by applying a series of ListenerConfigurers.
func (b *ListenerBuilder) Build() (envoy.NamedResource, error) {
	switch b.apiVersion {
	case envoy.APIV3:
		listener := envoy_listener_v3.Listener{}
		for _, configurer := range b.config.ConfigurersV3 {
			if err := configurer.Configure(&listener); err != nil {
				return nil, err
			}
		}
		return &listener, nil
	default:
		return nil, errors.New("unknown API")
	}
}

func (b *ListenerBuilder) MustBuild() envoy.NamedResource {
	listener, err := b.Build()
	if err != nil {
		panic(errors.Wrap(err, "failed to build Envoy Listener").Error())
	}

	return listener
}

// ListenerBuilderConfig holds configuration of a ListenerBuilder.
type ListenerBuilderConfig struct {
	// A series of ListenerConfigurers to apply to Envoy listener.
	ConfigurersV3 []v3.ListenerConfigurer
}

// AddV3 appends a given ListenerConfigurer to the end of the chain.
func (c *ListenerBuilderConfig) AddV3(configurer v3.ListenerConfigurer) {
	c.ConfigurersV3 = append(c.ConfigurersV3, configurer)
}

// ListenerBuilderOptFunc is a convenience type adapter.
type ListenerBuilderOptFunc func(config *ListenerBuilderConfig)

func (f ListenerBuilderOptFunc) ApplyTo(config *ListenerBuilderConfig) {
	if f != nil {
		f(config)
	}
}

// AddListenerConfigurer produces an option that applies the given
// configurer to the listener.
func AddListenerConfigurer(c v3.ListenerConfigurer) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.AddV3(c)
	})
}
