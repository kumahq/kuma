package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type StatsConfigurer struct {
	StatPrefix string
}

func (c *StatsConfigurer) Configure(l *envoy_listener.Listener) error {
	if c.StatPrefix != "" {
		l.StatPrefix = c.StatPrefix
	}

	return nil
}
