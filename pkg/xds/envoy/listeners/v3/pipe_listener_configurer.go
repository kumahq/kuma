package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type PipeListenerConfigurer struct {
	SocketPath string
}

func (c *PipeListenerConfigurer) Configure(l *envoy_api.Listener) error {
	l.Address = &envoy_core.Address{
		Address: &envoy_core.Address_Pipe{
			Pipe: &envoy_core.Pipe{
				Path: c.SocketPath,
			},
		},
	}

	return nil
}
