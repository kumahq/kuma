package mesh

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var defaultCircuitBreakerResource = func() model.Resource {
	return &core_mesh.CircuitBreakerResource{
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
}

var defaultMeshCircuitBreakerResource = func() model.Resource {
	return &v1alpha1.MeshCircuitBreakerResource{
		Spec: &v1alpha1.MeshCircuitBreaker{
			TargetRef: common_api.TargetRef{
				Kind: common_api.Mesh,
			},
			To: []v1alpha1.To{
				{
					TargetRef: common_api.TargetRef{
						Kind: common_api.Mesh,
					},
					Default: v1alpha1.Conf{
						ConnectionLimits: &v1alpha1.ConnectionLimits{
							MaxConnections:     pointer.To[uint32](1024),
							MaxPendingRequests: pointer.To[uint32](1024),
							MaxRequests:        pointer.To[uint32](1024),
							MaxRetries:         pointer.To[uint32](3),
						},
					},
				},
			},
		},
	}
}
