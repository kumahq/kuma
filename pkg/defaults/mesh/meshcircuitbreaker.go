package mesh

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

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
						OutlierDetection: &v1alpha1.OutlierDetection{
							Disabled:  pointer.To[bool](false),
							Detectors: &v1alpha1.Detectors{
								TotalFailures: &v1alpha1.DetectorTotalFailures{},
								GatewayFailures: &v1alpha1.DetectorGatewayFailures{},
								LocalOriginFailures: &v1alpha1.DetectorLocalOriginFailures{},
							},
						},
					},
				},
			},
		},
	}
}
