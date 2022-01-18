package mesh

import (
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var defaultTimeoutResource = &core_mesh.TimeoutResource{
	Spec: &mesh_proto.Timeout{
		Sources: []*mesh_proto.Selector{{
			Match: mesh_proto.MatchAnyService(),
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: mesh_proto.MatchAnyService(),
		}},
		Conf: &mesh_proto.Timeout_Conf{
			ConnectTimeout: util_proto.Duration(5 * time.Second),
			Tcp: &mesh_proto.Timeout_Conf_Tcp{
				IdleTimeout: util_proto.Duration(1 * time.Hour),
			},
			Http: &mesh_proto.Timeout_Conf_Http{
				IdleTimeout:    util_proto.Duration(1 * time.Hour),
				RequestTimeout: util_proto.Duration(15 * time.Second),
			},
			Grpc: &mesh_proto.Timeout_Conf_Grpc{
				StreamIdleTimeout: util_proto.Duration(5 * time.Minute),
			},
		},
	},
}
