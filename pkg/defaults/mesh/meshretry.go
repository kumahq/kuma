package mesh

import (
	"time"

	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

var defaultMeshRetryResource = func() model.Resource {
	return &v1alpha1.MeshRetryResource{
		Spec: &v1alpha1.MeshRetry{
			TargetRef: &common_api.TargetRef{
				Kind: common_api.Mesh,
			},
			To: []v1alpha1.To{
				{
					TargetRef: common_api.TargetRef{
						Kind: common_api.Mesh,
					},
					Default: v1alpha1.Conf{
						TCP: &v1alpha1.TCP{
							MaxConnectAttempt: pointer.To[uint32](5),
						},
						HTTP: &v1alpha1.HTTP{
							NumRetries: pointer.To[uint32](5),
							PerTryTimeout: &kube_meta.Duration{
								Duration: 16 * time.Second,
							},
							BackOff: &v1alpha1.BackOff{
								BaseInterval: &kube_meta.Duration{
									Duration: 25 * time.Millisecond,
								},
								MaxInterval: &kube_meta.Duration{
									Duration: 250 * time.Millisecond,
								},
							},
						},
						GRPC: &v1alpha1.GRPC{
							NumRetries: pointer.To[uint32](5),
							PerTryTimeout: &kube_meta.Duration{
								Duration: 16 * time.Second,
							},
							BackOff: &v1alpha1.BackOff{
								BaseInterval: &kube_meta.Duration{
									Duration: 25 * time.Millisecond,
								},
								MaxInterval: &kube_meta.Duration{
									Duration: 250 * time.Millisecond,
								},
							},
						},
					},
				},
			},
		},
	}
}
