package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type TransparentProxyingConfigurer struct{}

func (c *TransparentProxyingConfigurer) Configure(l *envoy_listener.Listener) error {
	l.BindToPort = util_proto.Bool(false)
	return nil
}
