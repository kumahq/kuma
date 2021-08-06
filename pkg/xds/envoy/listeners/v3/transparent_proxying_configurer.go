package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type TransparentProxyingConfigurer struct {
}

func (c *TransparentProxyingConfigurer) Configure(l *envoy_listener.Listener) error {
	l.BindToPort = &wrapperspb.BoolValue{Value: false}
	return nil
}
