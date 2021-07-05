package v3

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type TransparentProxyingConfigurer struct {
}

func (c *TransparentProxyingConfigurer) Configure(l *envoy_listener.Listener) error {
	l.BindToPort = &wrapperspb.BoolValue{Value: false}
	return nil
}
