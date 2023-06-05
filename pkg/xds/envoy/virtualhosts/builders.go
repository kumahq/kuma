package virtualhosts

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"github.com/kumahq/kuma/pkg/xds/envoy"
)

type VirtualHostConfigurer interface {
	Configure(*envoy_route.VirtualHost) error
}

type VirtualHostBuilder struct {
	configurers []VirtualHostConfigurer
}

func (r *VirtualHostBuilder) Configure(
	configurers ...VirtualHostConfigurer,
) *VirtualHostBuilder {
	r.configurers = append(r.configurers, configurers...)

	return r
}

func (r *VirtualHostBuilder) Build() (envoy.NamedResource, error) {
	virtualHost := &envoy_route.VirtualHost{}

	for _, c := range r.configurers {
		if err := c.Configure(virtualHost); err != nil {
			return nil, err
		}
	}

	return virtualHost, nil
}

type VirtualHostConfigureFunc func(*envoy_route.VirtualHost) error

func (f VirtualHostConfigureFunc) Configure(r *envoy_route.VirtualHost) error {
	if f != nil {
		return f(r)
	}

	return nil
}

type VirtualHostMustConfigureFunc func(*envoy_route.VirtualHost)

func (f VirtualHostMustConfigureFunc) Configure(
	r *envoy_route.VirtualHost,
) error {
	if f != nil {
		f(r)
	}

	return nil
}
