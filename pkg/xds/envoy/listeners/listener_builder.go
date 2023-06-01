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
	ApplyTo(builder *ListenerBuilder)
}

func NewListenerBuilder(apiVersion core_xds.APIVersion) *ListenerBuilder {
	return &ListenerBuilder{
		apiVersion: apiVersion,
	}
}

// ListenerBuilder is responsible for generating an Envoy listener
// by applying a series of ListenerConfigurers.
type ListenerBuilder struct {
	apiVersion  core_xds.APIVersion
	configurers []v3.ListenerConfigurer
}

// Configure configures ListenerBuilder by adding individual ListenerConfigurers.
func (b *ListenerBuilder) Configure(opts ...ListenerBuilderOpt) *ListenerBuilder {
	for _, opt := range opts {
		opt.ApplyTo(b)
	}

	return b
}

// Build generates an Envoy listener by applying a series of ListenerConfigurers.
func (b *ListenerBuilder) Build() (envoy.NamedResource, error) {
	switch b.apiVersion {
	case envoy.APIV3:
		listener := envoy_listener_v3.Listener{}
		for _, configurer := range b.configurers {
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

// AddConfigurer appends a given ListenerConfigurer to the end of the chain.
func (b *ListenerBuilder) AddConfigurer(configurer v3.ListenerConfigurer) {
	b.configurers = append(b.configurers, configurer)
}

// ListenerBuilderOptFunc is a convenience type adapter.
type ListenerBuilderOptFunc func(builder *ListenerBuilder)

func (f ListenerBuilderOptFunc) ApplyTo(builder *ListenerBuilder) {
	if f != nil {
		f(builder)
	}
}

// AddListenerConfigurer produces an option that applies the given
// configurer to the listener.
func AddListenerConfigurer(c v3.ListenerConfigurer) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(builder *ListenerBuilder) {
		builder.AddConfigurer(c)
	})
}
