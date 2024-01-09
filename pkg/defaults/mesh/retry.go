package mesh

import (
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var defaultRetryResource = func() model.Resource {
	return &core_mesh.RetryResource{
		Spec: &mesh_proto.Retry{
			Sources: []*mesh_proto.Selector{
				{
					Match: map[string]string{
						mesh_proto.ServiceTag: "*",
					},
				},
			},
			Destinations: []*mesh_proto.Selector{
				{
					Match: map[string]string{
						mesh_proto.ServiceTag: "*",
					},
				},
			},
			Conf: &mesh_proto.Retry_Conf{
				Http: &mesh_proto.Retry_Conf_Http{
					NumRetries:    util_proto.UInt32(5),
					PerTryTimeout: util_proto.Duration(16 * time.Second),
					BackOff: &mesh_proto.Retry_Conf_BackOff{
						BaseInterval: util_proto.Duration(25 * time.Millisecond),
						MaxInterval:  util_proto.Duration(250 * time.Millisecond),
					},
				},
				Tcp: &mesh_proto.Retry_Conf_Tcp{
					MaxConnectAttempts: 5,
				},
				Grpc: &mesh_proto.Retry_Conf_Grpc{
					NumRetries:    util_proto.UInt32(5),
					PerTryTimeout: util_proto.Duration(16 * time.Second),
					BackOff: &mesh_proto.Retry_Conf_BackOff{
						BaseInterval: util_proto.Duration(25 * time.Millisecond),
						MaxInterval:  util_proto.Duration(250 * time.Millisecond),
					},
				},
			},
		},
	}
}

var defaultMeshRetryResource = func() model.Resource {
	return &v1alpha1.MeshRetryResource{
		Spec: &v1alpha1.MeshRetry{
			TargetRef: common_api.TargetRef{
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
							PerTryTimeout: &v1.Duration{
								Duration: 16 * time.Second,
							},
							BackOff: &v1alpha1.BackOff{
								BaseInterval: &v1.Duration{
									Duration: 25 * time.Millisecond,
								},
								MaxInterval: &v1.Duration{
									Duration: 250 * time.Millisecond,
								},
							},
						},
						GRPC: &v1alpha1.GRPC{
							NumRetries: pointer.To[uint32](5),
							PerTryTimeout: &v1.Duration{
								Duration: 16 * time.Second,
							},
							BackOff: &v1alpha1.BackOff{
								BaseInterval: &v1.Duration{
									Duration: 25 * time.Millisecond,
								},
								MaxInterval: &v1.Duration{
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
