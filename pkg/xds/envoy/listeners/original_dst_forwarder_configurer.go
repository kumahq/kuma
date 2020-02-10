package listeners

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func OriginalDstForwarder() ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&OriginalDstForwarderConfigurer{})
	})
}

type OriginalDstForwarderConfigurer struct {
}

func (c *OriginalDstForwarderConfigurer) Configure(l *v2.Listener) error {
	// TODO(yskopets): What is the up-to-date alternative ?
	l.UseOriginalDst = &wrappers.BoolValue{Value: true}

	return nil
}
