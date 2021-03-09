package v2

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
)

type CommonVirtualHostConfigurer struct {
	Name string
}

func (c CommonVirtualHostConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	virtualHost.Name = c.Name
	virtualHost.Domains = []string{"*"}
	return nil
}
