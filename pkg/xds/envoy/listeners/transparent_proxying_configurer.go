package listeners

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

func TransparentProxying(transparentProxying *mesh_proto.Dataplane_Networking_TransparentProxying) ListenerBuilderOpt {
	virtual := transparentProxying.GetRedirectPort() != 0
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		if virtual {
			config.Add(&TransparentProxyingConfigurer{})
		}
	})
}

type TransparentProxyingConfigurer struct {
}

func (c *TransparentProxyingConfigurer) Configure(l *v2.Listener) error {
	// TODO(yskopets): What is the up-to-date alternative ?
	l.DeprecatedV1 = &v2.Listener_DeprecatedV1{
		BindToPort: &wrappers.BoolValue{Value: false},
	}

	return nil
}
