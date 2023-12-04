package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	policies_defaults "github.com/kumahq/kuma/pkg/plugins/policies/core/defaults"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var DefaultTimeoutResource = func() model.Resource {
	return &core_mesh.TimeoutResource{
		Spec: &mesh_proto.Timeout{
			Sources: []*mesh_proto.Selector{{
				Match: mesh_proto.MatchAnyService(),
			}},
			Destinations: []*mesh_proto.Selector{{
				Match: mesh_proto.MatchAnyService(),
			}},
			Conf: &mesh_proto.Timeout_Conf{
				ConnectTimeout: util_proto.Duration(policies_defaults.DefaultConnectTimeout),
				Tcp: &mesh_proto.Timeout_Conf_Tcp{
					IdleTimeout: util_proto.Duration(policies_defaults.DefaultIdleTimeout),
				},
				Http: &mesh_proto.Timeout_Conf_Http{
					IdleTimeout:       util_proto.Duration(policies_defaults.DefaultIdleTimeout),
					RequestTimeout:    util_proto.Duration(policies_defaults.DefaultRequestTimeout),
					StreamIdleTimeout: util_proto.Duration(policies_defaults.DefaultStreamIdleTimeout),
				},
			},
		},
	}
}

// DefaultInboundTimeout returns timeouts for the inbound side. This resource is not created
// in the store. It's used directly in InboundProxyGenerator. In the future, it could be replaced
// with a new InboundTimeout policy. The main idea around these values is to have them either
// bigger than outbound side timeouts or disabled.
var DefaultInboundTimeout = func() *mesh_proto.Timeout_Conf {
	const factor = 2

	return &mesh_proto.Timeout_Conf{
		ConnectTimeout: util_proto.Duration(factor * policies_defaults.DefaultConnectTimeout),
		Tcp: &mesh_proto.Timeout_Conf_Tcp{
			IdleTimeout: util_proto.Duration(factor * policies_defaults.DefaultIdleTimeout),
		},
		Http: &mesh_proto.Timeout_Conf_Http{
			RequestTimeout:    util_proto.Duration(0),
			IdleTimeout:       util_proto.Duration(factor * policies_defaults.DefaultIdleTimeout),
			StreamIdleTimeout: util_proto.Duration(factor * policies_defaults.DefaultStreamIdleTimeout),
			MaxStreamDuration: util_proto.Duration(0),
		},
	}
}
