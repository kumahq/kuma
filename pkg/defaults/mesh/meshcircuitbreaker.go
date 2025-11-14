package mesh

import (
	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

var defaultMeshCircuitBreakerResource = func() model.Resource {
	return &v1alpha1.MeshCircuitBreakerResource{
		Spec: &v1alpha1.MeshCircuitBreaker{
			TargetRef: &common_api.TargetRef{
				Kind: common_api.Mesh,
			},
			To: &[]v1alpha1.To{
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
