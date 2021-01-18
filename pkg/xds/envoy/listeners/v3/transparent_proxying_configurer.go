package v3

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type TransparentProxyingConfigurer struct {
}

func (c *TransparentProxyingConfigurer) Configure(l *envoy_listener.Listener) error {
	l.BindToPort = &wrappers.BoolValue{Value: false}
	return nil
}
