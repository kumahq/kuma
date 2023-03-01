package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type OriginalDstForwarderConfigurer struct{}

var _ ListenerConfigurer = &OriginalDstForwarderConfigurer{}

func (c *OriginalDstForwarderConfigurer) Configure(l *envoy_listener.Listener) error {
	// TODO(yskopets): What is the up-to-date alternative ?
	l.UseOriginalDst = util_proto.Bool(true)
	return nil
}
