package listeners

import (
	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// ListenerBuilder is responsible for generating an Envoy listener
// by applying a series of ListenerConfigurers.
type ListenerBuilder struct {
	apiVersion  core_xds.APIVersion
	configurers []v3.ListenerConfigurer
	name        string
}

// ListenerBuilderOpt is a configuration option for ListenerBuilder.
//
// The goal of ListenerBuilderOpt is to facilitate fluent ListenerBuilder API.
type ListenerBuilderOpt interface {
	// ApplyTo adds ListenerConfigurer(s) to the ListenerBuilder.
	ApplyTo(builder *ListenerBuilder)
}

func NewListenerBuilder(apiVersion core_xds.APIVersion, name string) *ListenerBuilder {
	return &ListenerBuilder{
		apiVersion: apiVersion,
		name:       name,
	}
}

// NewInboundListenerBuilder creates an Inbound ListenBuilder
// with a default name: inbound:address:port
func NewInboundListenerBuilder(
	apiVersion core_xds.APIVersion,
	address string,
	port uint32,
	protocol core_xds.SocketAddressProtocol,
) *ListenerBuilder {
	listenerName := envoy_names.GetInboundListenerName(address, port)

	return NewListenerBuilder(apiVersion, listenerName).
		Configure(InboundListener(address, port, protocol))
}

// NewOutboundListenerBuilder creates an Outbound ListenBuilder
// with a default name: outbound:address:port
func NewOutboundListenerBuilder(
	apiVersion core_xds.APIVersion,
	address string,
	port uint32,
	protocol core_xds.SocketAddressProtocol,
) *ListenerBuilder {
	if address == "" {
		address = "127.0.0.1"
	}
	listenerName := envoy_names.GetOutboundListenerName(address, port)

	return NewListenerBuilder(apiVersion, listenerName).
		Configure(OutboundListener(address, port, protocol))
}

func (b *ListenerBuilder) WithOverwriteName(name string) *ListenerBuilder {
	b.name = name
	return b
}

func (b *ListenerBuilder) GetName() string {
	return b.name
}

func (b *ListenerBuilder) Configure(opts ...ListenerBuilderOpt) *ListenerBuilder {
	for _, opt := range opts {
		opt.ApplyTo(b)
	}

	return b
}

func (b *ListenerBuilder) Build() (envoy.NamedResource, error) {
	if b.apiVersion != envoy.APIV3 {
		return nil, errors.New("unknown API")
	}

	listener := envoy_listener_v3.Listener{
		Name: b.name,
	}

	for _, configurer := range b.configurers {
		if err := configurer.Configure(&listener); err != nil {
			return nil, err
		}
	}

	if listener.GetName() == "" {
		return nil, errors.New("listener name is required, but it was not provided")
	}

	return &listener, nil
}

func (b *ListenerBuilder) MustBuild() envoy.NamedResource {
	listener, err := b.Build()
	if err != nil {
		panic(errors.Wrap(err, "failed to build Envoy Listener").Error())
	}

	return listener
}

type listenerBuilderOptFunc func(builder *ListenerBuilder)

func (f listenerBuilderOptFunc) ApplyTo(builder *ListenerBuilder) {
	if f != nil {
		f(builder)
	}
}

func AddListenerConfigurer(c v3.ListenerConfigurer) ListenerBuilderOpt {
	return listenerBuilderOptFunc(func(builder *ListenerBuilder) {
		builder.configurers = append(builder.configurers, c)
	})
}
