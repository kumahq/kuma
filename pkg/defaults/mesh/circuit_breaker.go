package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var defaultCircuitBreakerResource = &core_mesh.CircuitBreakerResource{
	Spec: &mesh_proto.CircuitBreaker{
		Sources: []*mesh_proto.Selector{{
			Match: mesh_proto.MatchAnyService(),
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: mesh_proto.MatchAnyService(),
		}},
		Conf: &mesh_proto.CircuitBreaker_Conf{
			Thresholds: &mesh_proto.CircuitBreaker_Conf_Thresholds{
				MaxConnections:     util_proto.UInt32(1024),
				MaxPendingRequests: util_proto.UInt32(1024),
				MaxRequests:        util_proto.UInt32(1024),
				MaxRetries:         util_proto.UInt32(3),
			},
		},
	},
}
