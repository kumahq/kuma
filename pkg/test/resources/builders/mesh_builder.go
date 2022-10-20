package builders

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type MeshBuilder struct {
	res *core_mesh.MeshResource
}

func Mesh() *MeshBuilder {
	return &MeshBuilder{
		res: &core_mesh.MeshResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
				Name: core_model.DefaultMesh,
			},
			Spec: &mesh_proto.Mesh{},
		},
	}
}

func (m *MeshBuilder) Build() *core_mesh.MeshResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}

func (m *MeshBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), m.Build(), store.CreateBy(m.Key()))
}

func (m *MeshBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(m.res.GetMeta())
}

func (m *MeshBuilder) WithName(name string) *MeshBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Name = name
	return m
}

func (m *MeshBuilder) WithEnabledMTLSBackend(name string) *MeshBuilder {
	if m.res.Spec.Mtls == nil {
		m.res.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	}
	m.res.Spec.Mtls.EnabledBackend = name
	return m
}

func (m *MeshBuilder) WithBuiltinMTLSBackend(name string) *MeshBuilder {
	m.res.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	return m.AddBuiltinMTLSBackend(name)
}

func (m *MeshBuilder) WithoutMTLSBackends() *MeshBuilder {
	m.res.Spec.Mtls.Backends = nil
	return m
}

func (m *MeshBuilder) AddBuiltinMTLSBackend(name string) *MeshBuilder {
	if m.res.Spec.Mtls == nil {
		m.res.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	}
	m.res.Spec.Mtls.Backends = append(m.res.Spec.Mtls.Backends, &mesh_proto.CertificateAuthorityBackend{
		Name: name,
		Type: "builtin",
	})
	return m
}

func (m *MeshBuilder) With(fn func(resource *core_mesh.MeshResource)) *MeshBuilder {
	fn(m.res)
	return m
}
