package builders

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	common_api "github.com/kumahq/kuma/pkg/plugins/policies/common/api/v1alpha1"
	meshtrace_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type MeshTraceBuilder struct {
	res *meshtrace_proto.MeshTraceResource
}

func MeshTrace() *MeshTraceBuilder {
	return &MeshTraceBuilder{
		res: &meshtrace_proto.MeshTraceResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "mtp-1",
			},
			Spec: &meshtrace_proto.MeshTrace{
				Default: meshtrace_proto.Conf{},
			},
		},
	}
}

func (m *MeshTraceBuilder) WithTargetRef(targetRef common_api.TargetRef) *MeshTraceBuilder {
	m.res.Spec.TargetRef = targetRef
	return m
}

func (m *MeshTraceBuilder) WithZipkinBackend(zipkin *meshtrace_proto.ZipkinBackend) *MeshTraceBuilder {
	m.res.Spec.Default.Backends = []meshtrace_proto.Backend{
		{
			Zipkin: zipkin,
		},
	}
	return m
}

func (m *MeshTraceBuilder) Build() *meshtrace_proto.MeshTraceResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}
