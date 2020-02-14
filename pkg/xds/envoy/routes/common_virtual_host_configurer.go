package routes

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
)

func CommonVirtualHost(name string) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(config *VirtualHostBuilderConfig) {
		config.Add(&CommonVirtualHostConfigurer{
			name: name,
		})
	})
}

type CommonVirtualHostConfigurer struct {
	name string
}

func (c CommonVirtualHostConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	virtualHost.Name = c.name
	virtualHost.Domains = []string{"*"}
	return nil
}
