package v2

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

type TransparentProxyingConfigurer struct {
}

func (c *TransparentProxyingConfigurer) Configure(l *envoy_api.Listener) error {
	// TODO(yskopets): What is the up-to-date alternative ?
	l.DeprecatedV1 = &envoy_api.Listener_DeprecatedV1{
		BindToPort: &wrappers.BoolValue{Value: false},
	}

	return nil
}
