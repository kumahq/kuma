package v3

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type CommonVirtualHostConfigurer struct {
	Name string
}

func (c CommonVirtualHostConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	virtualHost.Name = c.Name
	virtualHost.Domains = []string{"*"}
	return nil
}
