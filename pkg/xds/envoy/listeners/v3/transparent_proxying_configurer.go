package v3

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type TransparentProxyingConfigurer struct {
}

func (c *TransparentProxyingConfigurer) Configure(l *envoy_listener.Listener) error {
	// TODO(yskopets): What is the up-to-date alternative ?
	l.DeprecatedV1 = &envoy_listener.Listener_DeprecatedV1{
		BindToPort: &wrappers.BoolValue{Value: false},
	}

	return nil
}
