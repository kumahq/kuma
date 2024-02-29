package builders

import (
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type RetryBuilder struct {
	res *core_mesh.RetryResource
}

func Retry() *RetryBuilder {
	return &RetryBuilder{
		res: &core_mesh.RetryResource{
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "retry-all-default",
			},
			Spec: &mesh_proto.Retry{
				Sources: []*mesh_proto.Selector{
					{
						Match: mesh_proto.TagSelector{
							"kuma.io/service": "*",
						},
					},
				},
				Destinations: []*mesh_proto.Selector{
					{
						Match: mesh_proto.TagSelector{
							"kuma.io/service": "*",
						},
					},
				},
				Conf: &mesh_proto.Retry_Conf{
					Tcp: &mesh_proto.Retry_Conf_Tcp{MaxConnectAttempts: 5},
					Http: &mesh_proto.Retry_Conf_Http{
						NumRetries:    util_proto.UInt32(5),
						PerTryTimeout: util_proto.Duration(time.Second * 16),
						BackOff: &mesh_proto.Retry_Conf_BackOff{
							BaseInterval: util_proto.Duration(time.Millisecond * 25),
							MaxInterval:  util_proto.Duration(time.Millisecond * 250),
						},
						RetriableStatusCodes: []uint32{500, 504},
					},
				},
			},
		},
	}
}

func (b *RetryBuilder) Build() *core_mesh.RetryResource {
	if err := b.res.Validate(); err != nil {
		panic(err)
	}
	return b.res
}
