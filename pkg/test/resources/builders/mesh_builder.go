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
	*core_mesh.MeshResource
}

func Mesh() *MeshBuilder {
	return &MeshBuilder{
		MeshResource: &core_mesh.MeshResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
				Name: core_model.DefaultMesh,
			},
			Spec: &mesh_proto.Mesh{},
		},
	}
}

func (m *MeshBuilder) Build() *core_mesh.MeshResource {
	if err := m.MeshResource.Validate(); err != nil {
		panic(err)
	}
	return m.MeshResource
}

func (m *MeshBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), m.Build(), store.CreateBy(m.Key()))
}

func (m *MeshBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(m.GetMeta())
}

func (m *MeshBuilder) WithName(name string) *MeshBuilder {
	m.Meta.(*test_model.ResourceMeta).Name = name
	return m
}

func (m *MeshBuilder) WithEnabledMTLSBackend(name string) *MeshBuilder {
	if m.Spec.Mtls == nil {
		m.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	}
	m.Spec.Mtls.EnabledBackend = name
	return m
}

func (m *MeshBuilder) WithBuiltinMTLSBackend(name string) *MeshBuilder {
	m.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	return m.AddBuiltinMTLSBackend(name)
}

func (m *MeshBuilder) AddBuiltinMTLSBackend(name string) *MeshBuilder {
	if m.Spec.Mtls == nil {
		m.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	}
	m.Spec.Mtls.Backends = append(m.Spec.Mtls.Backends, &mesh_proto.CertificateAuthorityBackend{
		Name: name,
		Type: "builtin",
	})
	return m
}
